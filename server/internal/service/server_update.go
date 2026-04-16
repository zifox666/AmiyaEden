package service

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"amiya-eden/global"

	"go.uber.org/zap"
)

// staticHtmlDir is the path where nginx serves frontend static files.
// In the single-container setup both nginx and the Go server run in the
// same container, so this path is directly accessible from Go.
const staticHtmlDir = "/usr/share/nginx/html"

// currentVersion is injected at build time via:
//
//	-ldflags "-X amiya-eden/internal/service.currentVersion=v1.2.3"
var currentVersion = "dev"

// ServerUpdateService handles GitHub Release self-upgrade.
type ServerUpdateService struct{}

func NewServerUpdateService() *ServerUpdateService {
	return &ServerUpdateService{}
}

// githubServerRelease maps the fields we care about from the GitHub Releases API.
type githubServerRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// CheckUpdateResponse is the DTO returned to the frontend.
type CheckUpdateResponse struct {
	CurrentVersion       string `json:"current_version"`
	LatestVersion        string `json:"latest_version"`
	HasUpdate            bool   `json:"has_update"`
	ReleaseNotes         string `json:"release_notes"`
	DownloadSize         int64  `json:"download_size"`
	FrontendDownloadSize int64  `json:"frontend_download_size"`
}

// CheckUpdate fetches the latest GitHub release and compares with the running version.
func (s *ServerUpdateService) CheckUpdate() (*CheckUpdateResponse, error) {
	release, err := s.fetchLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("获取最新版本信息失败: %w", err)
	}

	assetName := s.assetName()
	var downloadSize int64
	var frontendDownloadSize int64
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadSize = a.Size
		}
		if a.Name == frontendAssetName {
			frontendDownloadSize = a.Size
		}
	}

	hasUpdate := release.TagName != currentVersion && currentVersion != "dev"
	// always show update when version is "dev" so devs can test
	if currentVersion == "dev" {
		hasUpdate = true
	}

	return &CheckUpdateResponse{
		CurrentVersion:       currentVersion,
		LatestVersion:        release.TagName,
		HasUpdate:            hasUpdate,
		ReleaseNotes:         release.Body,
		DownloadSize:         downloadSize,
		FrontendDownloadSize: frontendDownloadSize,
	}, nil
}

// PerformUpgrade downloads the new binary, swaps it with the running one, then exits.
// Docker's restart: unless-stopped will restart the container with the new binary.
func (s *ServerUpdateService) PerformUpgrade() error {
	release, err := s.fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("获取最新版本信息失败: %w", err)
	}

	assetName := s.assetName()
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("在 Release %s 中未找到适用于当前平台的二进制文件: %s", release.TagName, assetName)
	}

	// Determine current binary path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取当前可执行文件路径: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("无法解析可执行文件路径: %w", err)
	}

	newPath := execPath + ".new"
	bakPath := execPath + ".bak"

	global.Logger.Info("[ServerUpdate] 开始下载新版本",
		zap.String("version", release.TagName),
		zap.String("url", downloadURL),
	)

	if err := s.downloadBinary(downloadURL, newPath); err != nil {
		_ = os.Remove(newPath)
		return fmt.Errorf("下载失败: %w", err)
	}

	// Make the new binary executable
	if err := os.Chmod(newPath, 0755); err != nil {
		_ = os.Remove(newPath)
		return fmt.Errorf("设置可执行权限失败: %w", err)
	}

	// Backup current binary
	_ = os.Remove(bakPath)
	if err := os.Rename(execPath, bakPath); err != nil {
		_ = os.Remove(newPath)
		return fmt.Errorf("备份当前二进制失败: %w", err)
	}

	// Swap new binary into place
	if err := os.Rename(newPath, execPath); err != nil {
		// Try to restore backup
		_ = os.Rename(bakPath, execPath)
		return fmt.Errorf("替换二进制失败: %w", err)
	}

	global.Logger.Info("[ServerUpdate] 二进制替换完成，准备重启",
		zap.String("version", release.TagName),
		zap.String("path", execPath),
	)

	// Exit after a short delay so the HTTP response can be sent to the client.
	// Docker restart: unless-stopped will bring the process back with the new binary.
	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	return nil
}

// frontendAssetName is the expected archive asset name in GitHub Releases.
const frontendAssetName = "amiya-eden-frontend-dist.tar.gz"

// PerformFrontendUpgrade downloads the frontend dist archive and extracts it
// into staticHtmlDir. nginx serves static files directly from disk so the
// new files take effect immediately without a reload.
func (s *ServerUpdateService) PerformFrontendUpgrade() error {
	release, err := s.fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("获取最新版本信息失败: %w", err)
	}

	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == frontendAssetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("在 Release %s 中未找到前端资源包: %s", release.TagName, frontendAssetName)
	}

	tmpFile := filepath.Join(os.TempDir(), "amiya-frontend-dist.tar.gz")

	global.Logger.Info("[ServerUpdate] 开始下载前端资源包",
		zap.String("version", release.TagName),
		zap.String("url", downloadURL),
	)

	if err := s.downloadBinary(downloadURL, tmpFile); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("下载前端资源包失败: %w", err)
	}
	defer os.Remove(tmpFile)

	if err := os.MkdirAll(staticHtmlDir, 0755); err != nil {
		return fmt.Errorf("创建静态文件目录失败: %w", err)
	}

	if err := extractTarGz(tmpFile, staticHtmlDir); err != nil {
		return fmt.Errorf("解压前端资源包失败: %w", err)
	}

	global.Logger.Info("[ServerUpdate] 前端资源包更新完成", zap.String("version", release.TagName))
	return nil
}

// extractTarGz extracts a .tar.gz archive into destDir.
// Each entry path is validated to prevent path traversal attacks.
func extractTarGz(src, destDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	cleanDest := filepath.Clean(destDir) + string(os.PathSeparator)
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		// Security: reject any path that escapes destDir
		target := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), cleanDest) &&
			filepath.Clean(target) != filepath.Clean(destDir) {
			return fmt.Errorf("非法路径被拒绝: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			// Limit copy size to prevent decompression bomb (1 GB should be enough for frontend dist)
			if _, err := io.Copy(out, io.LimitReader(tr, 1<<30)); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
	return nil
}

// assetName returns the expected binary asset filename for the current platform.
// GitHub Actions must produce a file with exactly this name.
func (s *ServerUpdateService) assetName() string {
	return fmt.Sprintf("amiya-eden-server-%s-%s", runtime.GOOS, runtime.GOARCH)
}

// fetchLatestRelease calls the GitHub API.
func (s *ServerUpdateService) fetchLatestRelease() (*githubServerRelease, error) {
	owner := global.Config.GitHub.Owner
	repo := global.Config.GitHub.Repo
	if owner == "" || repo == "" {
		return nil, errors.New("GitHub owner/repo 未配置，请在 config.yaml 中设置 github.owner 和 github.repo")
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var release githubServerRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

// downloadBinary downloads a file from url and writes it to destPath.
func (s *ServerUpdateService) downloadBinary(url, destPath string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载返回 %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
