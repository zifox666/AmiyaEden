package service

import (
	"amiya-eden/global"
	"amiya-eden/internal/model"
	"amiya-eden/internal/repository"
	"archive/zip"
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"context"
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
		global.Logger.Error("[SDE] 查询版本记录失败", zap.String("version", version), zap.Error(err))
		return false, version, fmt.Errorf("查询版本记录失败: %w", err)
	}
	if exists {
		global.Logger.Info("[SDE] 当前版本已是最新", zap.String("version", version))
		return false, version, nil
	}

	global.Logger.Info("[SDE] 发现新版本，开始更新", zap.String("version", version))
	if err := s.doImport(release); err != nil {
		global.Logger.Error("[SDE] 更新失败", zap.String("version", version), zap.Error(err))
		return false, version, fmt.Errorf("导入 SDE 失败: %w", err)
	}

	if err := s.repo.CreateVersion(&model.SdeVersion{
		Version: version,
		Note:    "auto import from " + release.TagName,
	}); err != nil {
		global.Logger.Error("[SDE] 记录版本信息失败", zap.String("version", version), zap.Error(err))
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

// doImport 找到 PostgreSQL SQL 资源并导入数据库
func (s *SdeService) doImport(release *githubRelease) error {
	// 优先匹配上游专用 PostgreSQL 格式文件（如 sde-postgres.sql.bz2）
	// 策略：postgres/pgsql 关键字 > 通用 sql 关键字；压缩格式优于纯 .sql
	var asset *githubAsset
	for i := range release.Assets {
		name := strings.ToLower(release.Assets[i].Name)
		isPostgres := strings.Contains(name, "postgres") || strings.Contains(name, "pgsql")
		isSql := strings.Contains(name, ".sql")
		if !isPostgres && !isSql {
			continue
		}
		candidate := &release.Assets[i]
		// 找到专用 postgres 压缩包则立即采用，终止搜索
		if isPostgres {
			asset = candidate
			break
		}
		// 通用 sql 作为后备，优先压缩格式
		if asset == nil {
			asset = candidate
		} else if strings.HasSuffix(name, ".bz2") || strings.HasSuffix(name, ".gz") {
			asset = candidate
		}
	}
	if asset == nil {
		return errors.New("未找到 PostgreSQL SQL 资源文件")
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

	// 导入到 PostgreSQL
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
	url := global.Config.SDE.DownloadURL
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

// importSQL 读取 PostgreSQL SQL dump 文件并执行
// 上游已提供批量 INSERT，直接按语句顺序在事务中执行即可：
//  1. 使用专用连接，确保 session_replication_role 全程生效
//  2. DDL（CREATE/DROP/ALTER/SET）自动提交当前事务后立即执行
//  3. DML 每 batchSize 条提交一次事务；DML 失败时立即回滚并开新事务继续
const batchSize = 200

func importSQL(sqlPath string) error {
	sqlDB, err := global.DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 占用一条专用连接，保证 session 级设置全程有效
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("获取专用连接失败: %w", err)
	}
	defer conn.Close()

	// 禁用触发器/外键约束检查，关闭同步提交以提速
	for _, pragma := range []string{
		"SET session_replication_role = 'replica'",
		"SET synchronous_commit = OFF",
	} {
		if _, e := conn.ExecContext(context.Background(), pragma); e != nil {
			global.Logger.Warn("[SDE] 设置优化参数失败", zap.String("sql", pragma), zap.Error(e))
		}
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SET session_replication_role = 'origin'")
		_, _ = conn.ExecContext(context.Background(), "SET synchronous_commit = ON")
	}()

	f, err := os.Open(sqlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)

	var stmt strings.Builder
	var stmtCount, errCount int

	// 开启第一个事务
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	txCount := 0

	// commitTx 提交当前事务并立即开启下一个
	commitTx := func() {
		if err := tx.Commit(); err != nil {
			global.Logger.Warn("[SDE] 事务提交失败，尝试回滚", zap.Error(err))
			_ = tx.Rollback()
		}
		tx, _ = conn.BeginTx(context.Background(), nil)
		txCount = 0
	}

	// commitFinalTx 仅提交最后一个事务，不再开启新事务（避免 conn.Close 时遗留 dangling transaction 阻塞连接池）
	commitFinalTx := func() {
		if err := tx.Commit(); err != nil {
			global.Logger.Warn("[SDE] 最终事务提交失败，尝试回滚", zap.Error(err))
			_ = tx.Rollback()
		}
	}

	// rollbackTx 回滚当前事务并立即开启下一个（DML 失败时调用）
	rollbackTx := func() {
		_ = tx.Rollback()
		tx, _ = conn.BeginTx(context.Background(), nil)
		txCount = 0
	}

	isDDL := func(upper string) bool {
		for _, kw := range []string{"CREATE ", "DROP ", "ALTER ", "TRUNCATE ", "SET "} {
			if strings.HasPrefix(upper, kw) {
				return true
			}
		}
		return false
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 跳过空行、注释，以及 dump 自带的事务控制语句（由我们自行管理）
		upperTrimmed := strings.ToUpper(trimmed)
		if trimmed == "" ||
			strings.HasPrefix(trimmed, "--") ||
			strings.HasPrefix(trimmed, "/*") ||
			upperTrimmed == "BEGIN;" || upperTrimmed == "BEGIN" ||
			upperTrimmed == "COMMIT;" || upperTrimmed == "COMMIT" ||
			upperTrimmed == "ROLLBACK;" || upperTrimmed == "ROLLBACK" {
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
				// DDL：先提交当前 DML 事务，然后在事务外直接执行
				commitTx()
				if _, execErr := conn.ExecContext(context.Background(), sql); execErr != nil {
					errCount++
					global.Logger.Warn("[SDE] DDL 执行失败，已跳过",
						zap.String("err", execErr.Error()),
						zap.String("sql_prefix", truncate(sql, 120)))
				} else {
					stmtCount++
				}
			} else {
				if _, execErr := tx.Exec(sql); execErr != nil {
					errCount++
					global.Logger.Warn("[SDE] DML 执行失败，回滚当前批次",
						zap.String("err", execErr.Error()),
						zap.String("sql_prefix", truncate(sql, 120)))
					// 立即回滚，避免事务进入 aborted 状态导致后续语句全部失败
					rollbackTx()
				} else {
					stmtCount++
					txCount++
					if txCount >= batchSize {
						commitTx()
					}
				}
			}
		}
	}

	// 提交剩余事务（最后一次不再开启新事务）
	commitFinalTx()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	global.Logger.Info("[SDE] SQL 导入完成",
		zap.Int("成功语句数", stmtCount),
		zap.Int("失败语句数", errCount))
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
