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
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// sdeDownloadDir 临时文件存放目录
const sdeDownloadDir = "tmp/sde"

const (
	sdeCheckMinInterval = 30 * time.Minute
	sdeUserAgent        = "AmiyaEden-SDE-Updater/1.0"
)

var sdeUpdateState = struct {
	mu              sync.Mutex
	lastAutoCheckAt time.Time
	lastSeenVersion string
}{}

// SdeService SDE 业务逻辑层
type SdeService struct {
	repo      *repository.SdeRepository
	sysConfig *repository.SysConfigRepository
}

func NewSdeService() *SdeService {
	return &SdeService{
		repo:      repository.NewSdeRepository(),
		sysConfig: repository.NewSysConfigRepository(),
	}
}

// 获取配置的辅助方法
func (s *SdeService) getAPIKey() string {
	key, _ := s.sysConfig.Get(model.SysConfigSDEAPIKey, global.Config.SDE.APIKey)
	return key
}

func (s *SdeService) getProxy() string {
	proxy, _ := s.sysConfig.Get(model.SysConfigSDEProxy, global.Config.SDE.Proxy)
	return proxy
}

func (s *SdeService) getDownloadURL() string {
	url, _ := s.sysConfig.Get(model.SysConfigSDEDownloadURL, global.Config.SDE.DownloadURL)
	return url
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
	sdeUpdateState.mu.Lock()
	defer sdeUpdateState.mu.Unlock()

	now := time.Now()
	if !sdeUpdateState.lastAutoCheckAt.IsZero() && now.Sub(sdeUpdateState.lastAutoCheckAt) < sdeCheckMinInterval {
		global.Logger.Info("[SDE] 跳过重复自动检查",
			zap.Duration("min_interval", sdeCheckMinInterval),
			zap.Time("last_check_at", sdeUpdateState.lastAutoCheckAt),
			zap.String("last_seen_version", sdeUpdateState.lastSeenVersion))
		return false, sdeUpdateState.lastSeenVersion, nil
	}
	sdeUpdateState.lastAutoCheckAt = now

	release, err := s.fetchLatestRelease()
	if err != nil {
		return false, "", fmt.Errorf("获取 GitHub Release 失败: %w", err)
	}

	version := release.TagName
	sdeUpdateState.lastSeenVersion = version
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
	sdeUpdateState.mu.Lock()
	defer sdeUpdateState.mu.Unlock()

	release, err := s.fetchLatestRelease()
	if err != nil {
		return "", fmt.Errorf("获取 GitHub Release 失败: %w", err)
	}

	version := release.TagName
	sdeUpdateState.lastSeenVersion = version
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
	if err := s.downloadFile(asset.BrowserDownloadURL, dlPath); err != nil {
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
func (s *SdeService) newHTTPClient(timeout time.Duration) *http.Client {
	return s.newHTTPClientWithProxy(timeout, true)
}

func (s *SdeService) newHTTPClientWithProxy(timeout time.Duration, useProxy bool) *http.Client {
	transport := &http.Transport{}
	if useProxy {
		if proxyAddr := s.getProxy(); proxyAddr != "" {
			if proxyURL, err := url.Parse(proxyAddr); err == nil {
				transport.Proxy = http.ProxyURL(proxyURL)
			} else {
				global.Logger.Warn("[SDE] 代理地址解析失败，将不使用代理", zap.String("proxy", proxyAddr), zap.Error(err))
			}
		}
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

func (s *SdeService) doRequestWithProxyFallback(timeout time.Duration, build func(*http.Client) (*http.Response, error)) (*http.Response, error) {
	client := s.newHTTPClientWithProxy(timeout, true)
	resp, err := build(client)
	if err == nil {
		return resp, nil
	}

	if !s.shouldRetryWithoutProxy(err) {
		return nil, err
	}

	global.Logger.Warn("[SDE] 代理请求失败，回退为直连重试", zap.Error(err))
	directClient := s.newHTTPClientWithProxy(timeout, false)
	return build(directClient)
}

func (s *SdeService) shouldRetryWithoutProxy(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "proxyconnect") ||
		strings.Contains(msg, "connect: connection refused") ||
		strings.Contains(msg, "socks")
}

// fetchLatestRelease 获取 GitHub 最新 release 信息
func (s *SdeService) fetchLatestRelease() (*githubRelease, error) {
	url := s.getDownloadURL()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", sdeUserAgent)
	resp, err := s.doRequestWithProxyFallback(30*time.Second, func(client *http.Client) (*http.Response, error) {
		return client.Do(req.Clone(context.Background()))
	})
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
func (s *SdeService) downloadFile(url, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", sdeUserAgent)

	resp, err := s.doRequestWithProxyFallback(10*time.Minute, func(client *http.Client) (*http.Response, error) {
		return client.Do(req.Clone(context.Background()))
	})
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
// 上游 SQL 以多行 INSERT ... VALUES (...) 形式组织大量数据（尤其是 trnTranslations）。
// 为避免超大单条 INSERT 导致执行失败或导入不完整，这里会按行流式拆分为较小块执行。
//
// 执行策略：
//  1. 使用专用连接，确保 session_replication_role 全程生效
//  2. DDL（CREATE/DROP/ALTER/TRUNCATE/SET）自动提交当前 DML 事务后立即执行
//  3. DML 每 batchSize 条提交一次事务
//  4. 多行 INSERT ... VALUES 每 insertChunkSize 行拆分为独立 INSERT 执行
//  5. DML 失败视为致命错误，立即中止，避免产生“成功但数据不完整”的导入结果
const (
	batchSize       = 200
	insertChunkSize = 1000
)

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
	var stmtCount, ddlErrCount int
	var insertHeader string
	var insertRows []string

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

	execDDL := func(sql string) {
		commitTx()
		if _, execErr := conn.ExecContext(context.Background(), sql); execErr != nil {
			ddlErrCount++
			global.Logger.Warn("[SDE] DDL 执行失败，已跳过",
				zap.String("err", execErr.Error()),
				zap.String("sql_prefix", truncate(sql, 120)))
			return
		}
		stmtCount++
	}

	execDML := func(sql string) error {
		if _, execErr := tx.Exec(sql); execErr != nil {
			global.Logger.Error("[SDE] DML 执行失败，导入终止",
				zap.String("err", execErr.Error()),
				zap.String("sql_prefix", truncate(sql, 160)))
			rollbackTx()
			return execErr
		}

		stmtCount++
		txCount++
		if txCount >= batchSize {
			commitTx()
		}
		return nil
	}

	normalizeInsertRow := func(row string, isLast bool) string {
		trimmed := strings.TrimSpace(row)
		trimmed = strings.TrimSuffix(trimmed, ",")
		trimmed = strings.TrimSuffix(trimmed, ";")
		if isLast {
			return trimmed + ";"
		}
		return trimmed + ","
	}

	flushInsertRows := func(final bool) error {
		if insertHeader == "" || len(insertRows) == 0 {
			if final {
				insertHeader = ""
				insertRows = nil
			}
			return nil
		}

		for len(insertRows) > 0 {
			chunkSize := insertChunkSize
			if len(insertRows) < chunkSize {
				chunkSize = len(insertRows)
			}

			rows := make([]string, chunkSize)
			for i := 0; i < chunkSize; i++ {
				rows[i] = normalizeInsertRow(insertRows[i], i == chunkSize-1)
			}

			sql := insertHeader + "\n" + strings.Join(rows, "\n")
			if err := execDML(sql); err != nil {
				return err
			}
			insertRows = insertRows[chunkSize:]
		}

		if final {
			insertHeader = ""
			insertRows = nil
		}
		return nil
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

		if insertHeader != "" {
			insertRows = append(insertRows, trimmed)

			if strings.HasSuffix(trimmed, ";") {
				if err := flushInsertRows(true); err != nil {
					return fmt.Errorf("执行批量 INSERT 失败: %w", err)
				}
				continue
			}

			if len(insertRows) >= insertChunkSize {
				if err := flushInsertRows(false); err != nil {
					return fmt.Errorf("执行批量 INSERT 失败: %w", err)
				}
			}
			continue
		}

		if upperTrimmed == "INSERT INTO" || (strings.HasPrefix(upperTrimmed, "INSERT INTO ") && strings.HasSuffix(upperTrimmed, " VALUES")) {
			insertHeader = trimmed
			insertRows = insertRows[:0]
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
				execDDL(sql)
			} else {
				if err := execDML(sql); err != nil {
					return fmt.Errorf("执行 SQL 失败: %w", err)
				}
			}
		}
	}

	if insertHeader != "" {
		return errors.New("SQL dump 以未结束的 INSERT 语句结尾")
	}

	// 提交剩余事务（最后一次不再开启新事务）
	commitFinalTx()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 SQL 文件失败: %w", err)
	}

	global.Logger.Info("[SDE] SQL 导入完成",
		zap.Int("成功语句数", stmtCount),
		zap.Int("DDL失败语句数", ddlErrCount))
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

// GetNames 批量查询按 namespace 分组的 id -> name 映射（仅查数据库翻译表）
func (s *SdeService) GetNames(ids map[string][]int, languageID string) (repository.SdeNameMap, error) {
	return s.repo.GetNames(ids, languageID)
}

// FuzzySearch 模糊搜索物品/成员名称
func (s *SdeService) FuzzySearch(keyword string, languageID string, categoryIDs []int, excludeCategoryIDs []int, limit int, searchMember bool) ([]repository.FuzzySearchItem, error) {
	return s.repo.FuzzySearch(keyword, languageID, categoryIDs, excludeCategoryIDs, limit, searchMember)
}
