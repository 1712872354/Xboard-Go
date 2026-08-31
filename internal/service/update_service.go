package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"xboard-go/pkg/logger"
)

// UpdateService 系统更新服务接口
type UpdateService interface {
	// CheckUpdate 检查更新
	CheckUpdate() (*UpdateInfo, error)
	// ExecuteUpdate 执行更新
	ExecuteUpdate() error
}

type updateService struct {
	currentVersion string
	githubRepo     string
	httpClient     *http.Client
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	HasUpdate    bool         `json:"has_update"`
	CurrentVersion string     `json:"current_version"`
	LatestVersion string      `json:"latest_version"`
	DownloadURL  string       `json:"download_url"`
	PublishedAt  string       `json:"published_at"`
	Author       string       `json:"author"`
	UpdateLogs   []UpdateLog  `json:"update_logs"`
}

// UpdateLog 更新日志
type UpdateLog struct {
	Version string `json:"version"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

// GitHubRelease GitHub Release 响应
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Author  struct {
		Login string `json:"login"`
	} `json:"author"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// NewUpdateService 创建更新服务
func NewUpdateService(currentVersion string) UpdateService {
	return &updateService{
		currentVersion: currentVersion,
		githubRepo:     "cedar2025/Xboard",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckUpdate 检查更新
func (s *updateService) CheckUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", s.githubRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Xboard-Go-Update-Checker")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 查找对应平台的下载链接
	downloadURL := s.findDownloadURL(release.Assets)

	hasUpdate := release.TagName != s.currentVersion && release.TagName != "v"+s.currentVersion

	info := &UpdateInfo{
		HasUpdate:     hasUpdate,
		CurrentVersion: s.currentVersion,
		LatestVersion: release.TagName,
		DownloadURL:   downloadURL,
		PublishedAt:   release.PublishedAt,
		Author:        release.Author.Login,
		UpdateLogs: []UpdateLog{
			{
				Version: release.TagName,
				Message: release.Body,
				Author:  release.Author.Login,
				Date:    release.PublishedAt,
			},
		},
	}

	return info, nil
}

// ExecuteUpdate 执行更新
func (s *updateService) ExecuteUpdate() error {
	// 1. 检查更新
	info, err := s.CheckUpdate()
	if err != nil {
		return fmt.Errorf("check update failed: %w", err)
	}

	if !info.HasUpdate {
		return fmt.Errorf("already latest version")
	}

	if info.DownloadURL == "" {
		return fmt.Errorf("no download URL available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// 2. 下载新版本
	tmpFile, err := s.downloadBinary(info.DownloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// 3. 备份当前版本
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get current executable failed: %w", err)
	}

	backupFile := currentExe + ".bak"
	if err := copyFile(currentExe, backupFile); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// 4. 替换二进制文件
	if err := copyFile(tmpFile, currentExe); err != nil {
		// 回滚
		copyFile(backupFile, currentExe)
		return fmt.Errorf("replace failed: %w", err)
	}

	// 5. 设置可执行权限
	if err := os.Chmod(currentExe, 0755); err != nil {
		return fmt.Errorf("chmod failed: %w", err)
	}

	logger.Sugar().Infof("Update completed: %s -> %s", s.currentVersion, info.LatestVersion)

	// 6. 重启服务（通过外部脚本或 systemd）
	// 这里只是标记更新完成，实际重启需要外部机制
	return nil
}

// findDownloadURL 查找对应平台的下载链接
func (s *updateService) findDownloadURL(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) string {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// 映射架构名
	switch archName {
	case "amd64":
		archName = "amd64"
	case "arm64":
		archName = "arm64"
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			return asset.BrowserDownloadURL
		}
	}

	// 尝试通用匹配
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

// downloadBinary 下载二进制文件
func (s *updateService) downloadBinary(url string) (string, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "xboard-update-*")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

// RestartService 重启服务
func RestartService() error {
	// 获取当前可执行文件路径
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// 使用 exec 重启
	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	// 退出当前进程
	os.Exit(0)
	return nil
}
