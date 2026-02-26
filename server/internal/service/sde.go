package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"archive/zip"
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// sdeDownloadDir 临时文件存放目录
const sdeDownloadDir = "tmp/sde"

// SdeService SDE 业务逻辑层
type SdeService struct {
	repo *repository.SdeRepository
}

func NewSdeService() *SdeService {
	return &SdeService{repo: repository.NewSdeRepository()}
}

// ---- 版本管理 ----

// GetCurrentVersion 获取当前已导入的 SDE 版本
func (s *SdeService) GetCurrentVersion() (*model.SdeVersion, error) {
	v, err := s.repo.GetLatestVersion()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return v, err
}

// ---- SDE 更新 ----

// githubRelease GitHub Releases API 响应结构（仅摘取需要的字段）
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckAndUpdate 检查最新 SDE 版本，若有新版本则自动下载并导入
// 返回 (isUpdated, version, error)
func (s *SdeService) CheckAndUpdate() (bool, string, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return false, "", fmt.Errorf("获取 GitHub Release 失败: %w", err)
	}

	version := release.TagName
	exists, err := s.repo.VersionExists(version)
	if err != nil {
		return false, version, fmt.Errorf("查询版本记录失败: %w", err)
	}
	if exists {
		global.Logger.Info("[SDE] 当前版本已是最新", zap.String("version", version))
		return false, version, nil
	}

	global.Logger.Info("[SDE] 发现新版本，开始更新", zap.String("version", version))
	if err := s.doImport(release); err != nil {
		return false, version, fmt.Errorf("导入 SDE 失败: %w", err)
	}

	if err := s.repo.CreateVersion(&model.SdeVersion{
		Version: version,
		Note:    "auto import from " + release.TagName,
	}); err != nil {
		return true, version, fmt.Errorf("记录版本失败: %w", err)
	}

	global.Logger.Info("[SDE] 更新完成", zap.String("version", version))
	return true, version, nil
}

// TriggerManualUpdate 手动触发更新，强制重新导入
func (s *SdeService) TriggerManualUpdate() (string, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return "", fmt.Errorf("获取 GitHub Release 失败: %w", err)
	}

	version := release.TagName
	global.Logger.Info("[SDE] 手动触发更新", zap.String("version", version))

	if err := s.doImport(release); err != nil {
		return version, fmt.Errorf("导入 SDE 失败: %w", err)
	}

	// 如果版本已存在就更新，否则插入
	exists, _ := s.repo.VersionExists(version)
	if !exists {
		_ = s.repo.CreateVersion(&model.SdeVersion{
			Version: version,
			Note:    "manual import",
		})
	}

	global.Logger.Info("[SDE] 手动更新完成", zap.String("version", version))
	return version, nil
}

// doImport 找到 SQL 资源并导入数据库
func (s *SdeService) doImport(release *githubRelease) error {
	// 找到 MySQL SQL 资源（优先 .sql.gz，其次 .zip，最后 .sql）
	var asset *githubAsset
	for i := range release.Assets {
		name := strings.ToLower(release.Assets[i].Name)
		if strings.Contains(name, "mysql") || strings.Contains(name, "sql") {
			asset = &release.Assets[i]
			if strings.HasSuffix(name, ".sql.gz") || strings.HasSuffix(name, ".gz") ||
				strings.HasSuffix(name, ".sql.bz2") || strings.HasSuffix(name, ".bz2") {
				break // 优先压缩格式
			}
		}
	}
	if asset == nil {
		return errors.New("未找到 SQL 资源文件")
	}

	// 确保临时目录存在
	if err := os.MkdirAll(sdeDownloadDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 下载
	dlPath := filepath.Join(sdeDownloadDir, asset.Name)
	global.Logger.Info("[SDE] 下载中", zap.String("url", asset.BrowserDownloadURL))
	if err := downloadFile(asset.BrowserDownloadURL, dlPath); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer os.Remove(dlPath)

	// 解压得到 .sql 文件路径
	sqlPath, err := extractSQL(dlPath, sdeDownloadDir)
	if err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}
	if sqlPath != dlPath {
		defer os.Remove(sqlPath)
	}

	// 导入到 MySQL
	global.Logger.Info("[SDE] 导入数据库中", zap.String("file", sqlPath))
	if err := importSQL(sqlPath); err != nil {
		return fmt.Errorf("导入数据库失败: %w", err)
	}

	return nil
}

// ---- 工具函数 ----

// newHTTPClient 根据配置创建 http.Client，若设置了代理则使用代理
func newHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{}
	if proxyAddr := global.Config.SDE.Proxy; proxyAddr != "" {
		if proxyURL, err := url.Parse(proxyAddr); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			global.Logger.Warn("[SDE] 代理地址解析失败，将不使用代理", zap.String("proxy", proxyAddr), zap.Error(err))
		}
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

// fetchLatestRelease 获取 GitHub 最新 release 信息
func fetchLatestRelease() (*githubRelease, error) {
	url := "https://api.github.com/repos/zifox666/eve-sde-converter/releases/latest"
	client := newHTTPClient(30 * time.Second)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回非 200: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// downloadFile 下载文件到指定路径
func downloadFile(url, destPath string) error {
	client := newHTTPClient(10 * time.Minute)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载返回非 200: %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// extractSQL 通过魔数检测文件类型并递归解压，最终返回 .sql 文件路径
// 支持：plain SQL / gzip(.gz/.sql.gz) / zip（内含 .sql 或 .sql.gz）
func extractSQL(srcPath, destDir string) (string, error) {
	magic, err := readMagicBytes(srcPath)
	if err != nil {
		return "", fmt.Errorf("读取文件魔数失败: %w", err)
	}

	switch {
	case isGzipMagic(magic):
		return extractGzip(srcPath, destDir)
	case isBzip2Magic(magic):
		return extractBzip2(srcPath, destDir)
	case isZipMagic(magic):
		return extractZip(srcPath, destDir)
	default:
		// 当作纯 SQL 文件处理
		return srcPath, nil
	}
}

// readMagicBytes 读取文件头 4 字节
func readMagicBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, 4)
	n, _ := f.Read(buf)
	return buf[:n], nil
}

// isGzipMagic 判断是否为 gzip 格式（魔数 1f 8b）
func isGzipMagic(magic []byte) bool {
	return len(magic) >= 2 && magic[0] == 0x1f && magic[1] == 0x8b
}

// isBzip2Magic 判断是否为 bzip2 格式（魔数 42 5a 68 = "BZh"）
func isBzip2Magic(magic []byte) bool {
	return len(magic) >= 3 && magic[0] == 0x42 && magic[1] == 0x5a && magic[2] == 0x68
}

// isZipMagic 判断是否为 zip 格式（魔数 50 4b 03 04）
func isZipMagic(magic []byte) bool {
	return len(magic) >= 4 && magic[0] == 0x50 && magic[1] == 0x4b &&
		magic[2] == 0x03 && magic[3] == 0x04
}

// extractGzip 解压 gzip 文件，输出文件名去掉 .gz 后缀
func extractGzip(srcPath, destDir string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	// gzip 头中存有原始文件名
	outName := gr.Header.Name
	if outName == "" {
		outName = strings.TrimSuffix(filepath.Base(srcPath), ".gz")
	}
	outPath := filepath.Join(destDir, filepath.Base(outName))
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, gr); err != nil {
		return "", err
	}

	// 解压结果可能还是压缩包，递归处理
	return extractSQL(outPath, destDir)
}

// extractBzip2 解压 bzip2 文件，输出文件名去掉 .bz2 后缀
func extractBzip2(srcPath, destDir string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	br := bzip2.NewReader(f)

	outName := strings.TrimSuffix(filepath.Base(srcPath), ".bz2")
	outPath := filepath.Join(destDir, outName)
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, br); err != nil {
		return "", err
	}

	// 解压结果可能还是压缩包，递归处理
	return extractSQL(outPath, destDir)
}

// extractZip 解压 zip，找到第一个 SQL 相关条目（.sql 或 .sql.gz）并递归解压
func extractZip(srcPath, destDir string) (string, error) {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		if !strings.Contains(name, ".sql") {
			continue
		}
		outPath := filepath.Join(destDir, filepath.Base(f.Name))
		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		out, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, copyErr := io.Copy(out, rc)
		out.Close()
		rc.Close()
		if copyErr != nil {
			return "", copyErr
		}

		// 递归：内部文件可能还是 gzip
		return extractSQL(outPath, destDir)
	}
	return "", errors.New("zip 中未找到 SQL 相关文件")
}

// importSQL 读取 SQL dump 文件并批量事务执行
// 优化点：
//  1. SET FOREIGN_KEY_CHECKS=0 / UNIQUE_CHECKS=0 关闭约束检查
//  2. DDL（CREATE/DROP/ALTER/LOCK/UNLOCK）立即执行，不进事务
//  3. DML（INSERT/UPDATE/DELETE/REPLACE）每 batchSize 条提交一次事务
//  4. 限制错误日志频率，只输出错误摘要
const batchSize = 500

// mergeInserts 尝试将多条 INSERT INTO t VALUES (...); 合并为一条多值 INSERT
// 如果语句格式不匹配则返回 ("", false)
func mergeInserts(stmts []string) (string, bool) {
	if len(stmts) == 0 {
		return "", false
	}
	if len(stmts) == 1 {
		return stmts[0], true
	}

	// 找第一条的 VALUES 位置，提取表前缀 "INSERT INTO t VALUES "
	first := stmts[0]
	upperFirst := strings.ToUpper(first)
	valIdx := strings.Index(upperFirst, " VALUES ")
	if valIdx < 0 {
		return "", false
	}
	prefix := first[:valIdx+8] // "INSERT INTO t VALUES "
	upperPrefix := strings.ToUpper(prefix)

	var sb strings.Builder
	sb.Grow(len(first) * len(stmts))
	// 写第一条（去掉末尾分号）
	body := strings.TrimSuffix(strings.TrimSpace(first), ";")
	sb.WriteString(body)

	for _, s := range stmts[1:] {
		upper := strings.ToUpper(s)
		// 必须是同一张表的 INSERT
		if !strings.HasPrefix(upper, upperPrefix) {
			return "", false
		}
		// 取 VALUES 后面的值部分
		valPart := strings.TrimSuffix(strings.TrimSpace(s[valIdx+8:]), ";")
		sb.WriteByte(',')
		sb.WriteString(valPart)
	}
	sb.WriteByte(';')
	return sb.String(), true
}

func importSQL(sqlPath string) error {
	sqlDB, err := global.DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 连接级别优化参数
	for _, pragma := range []string{
		"SET FOREIGN_KEY_CHECKS=0",
		"SET UNIQUE_CHECKS=0",
		"SET SQL_LOG_BIN=0",
		"SET autocommit=0",
	} {
		if _, e := sqlDB.Exec(pragma); e != nil {
			global.Logger.Warn("[SDE] 设置优化参数失败", zap.String("sql", pragma), zap.Error(e))
		}
	}
	defer func() {
		for _, pragma := range []string{
			"COMMIT",
			"SET FOREIGN_KEY_CHECKS=1",
			"SET UNIQUE_CHECKS=1",
			"SET autocommit=1",
		} {
			_, _ = sqlDB.Exec(pragma)
		}
	}()

	f, err := os.Open(sqlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)

	var stmt strings.Builder
	var stmtCount, errCount int

	// DML 批次：key=INSERT前缀(表名), value=该表的待合并语句
	// 用 slice 保序，map 记录位置
	type batchEntry struct {
		prefix string
		stmts  []string
	}
	var batchOrder []string                  // 保持插入顺序的前缀列表
	batchMap := make(map[string]*batchEntry) // prefix -> entry

	// flushTable 合并并执行某张表的积累语句
	flushTable := func(entry *batchEntry) {
		if len(entry.stmts) == 0 {
			return
		}
		merged, ok := mergeInserts(entry.stmts)
		if !ok {
			// 合并失败，逐条执行
			for _, s := range entry.stmts {
				if _, e := sqlDB.Exec(s); e != nil {
					errCount++
				} else {
					stmtCount++
				}
			}
			return
		}
		if _, e := sqlDB.Exec(merged); e != nil {
			// 合并失败则逐条重试
			for _, s := range entry.stmts {
				if _, e2 := sqlDB.Exec(s); e2 != nil {
					errCount++
				} else {
					stmtCount++
				}
			}
		} else {
			stmtCount += len(entry.stmts)
		}
	}

	// commitBatch 提交所有积累的 DML，按表顺序 flush，然后 COMMIT
	commitBatch := func() {
		for _, prefix := range batchOrder {
			flushTable(batchMap[prefix])
		}
		batchOrder = batchOrder[:0]
		for k := range batchMap {
			delete(batchMap, k)
		}
		_, _ = sqlDB.Exec("COMMIT")
		_, _ = sqlDB.Exec("SET autocommit=0")
	}

	isDDL := func(upper string) bool {
		for _, prefix := range []string{
			"CREATE ", "DROP ", "ALTER ", "TRUNCATE ",
			"LOCK ", "UNLOCK ", "SET ", "USE ",
		} {
			if strings.HasPrefix(upper, prefix) {
				return true
			}
		}
		return false
	}

	// 获取 INSERT 语句的表前缀（用于分组合并）
	getInsertPrefix := func(sql string) string {
		upper := strings.ToUpper(sql)
		if !strings.HasPrefix(upper, "INSERT") {
			return ""
		}
		valIdx := strings.Index(upper, " VALUES ")
		if valIdx < 0 {
			return ""
		}
		return strings.ToUpper(sql[:valIdx+8])
	}

	totalDML := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" ||
			strings.HasPrefix(trimmed, "--") ||
			strings.HasPrefix(trimmed, "#") ||
			(strings.HasPrefix(trimmed, "/*!") && strings.HasSuffix(trimmed, "*/;")) ||
			(strings.HasPrefix(trimmed, "/*!") && strings.HasSuffix(trimmed, "*/")) {
			continue
		}

		stmt.WriteString(line)
		stmt.WriteByte('\n')

		if strings.HasSuffix(trimmed, ";") {
			sql := strings.TrimSpace(stmt.String())
			stmt.Reset()
			if sql == "" || sql == ";" {
				continue
			}

			upper := strings.ToUpper(sql)
			if isDDL(upper) {
				commitBatch()
				if _, execErr := sqlDB.Exec(sql); execErr != nil {
					errCount++
					global.Logger.Warn("[SDE] DDL 执行失败，已跳过",
						zap.String("err", execErr.Error()),
						zap.String("sql_prefix", truncate(sql, 120)))
				} else {
					stmtCount++
				}
			} else {
				prefix := getInsertPrefix(sql)
				if prefix == "" {
					// 非 INSERT DML，直接加入通用批次
					prefix = "__other__"
				}
				entry, exists := batchMap[prefix]
				if !exists {
					entry = &batchEntry{prefix: prefix}
					batchMap[prefix] = entry
					batchOrder = append(batchOrder, prefix)
				}
				entry.stmts = append(entry.stmts, sql)
				totalDML++

				if totalDML >= batchSize {
					commitBatch()
					totalDML = 0
				}
			}
		}
	}

	commitBatch()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	global.Logger.Info("[SDE] SQL 导入完成",
		zap.Int("成功语句数", stmtCount),
		zap.Int("跳过语句数", errCount))
	return nil
}

// truncate 截断字符串，用于日志输出
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ---- 数据查询 ----

// GetTypes 批量查询物品信息（含 group + category + market_group 翻译）
func (s *SdeService) GetTypes(typeIDs []int, published *bool, languageID string) ([]repository.TypeInfo, error) {
	return s.repo.GetTypes(typeIDs, published, languageID)
}

// GetNames 批量查询 id -> name 映射（仅查数据库翻译表）
func (s *SdeService) GetNames(ids map[string][]int, languageID string) (map[int]string, error) {
	return s.repo.GetNames(ids, languageID)
}
