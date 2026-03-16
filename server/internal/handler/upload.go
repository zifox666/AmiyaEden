package handler

import (
	"amiya-eden/pkg/response"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	uploadDir     = "./uploads"
	maxUploadSize = 5 << 20 // 5 MB
)

var allowedExts = map[string]bool{
	".jpg": true,
	".jpeg": true,
	".png": true,
	".gif": true,
	".webp": true,
}

// UploadHandler 文件上传处理器
type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	// 确保上传目录存在
	_ = os.MkdirAll(uploadDir, 0755)
	return &UploadHandler{}
}

// UploadImage POST /api/v1/upload/image
// 上传图片，返回可访问的 URL
func (h *UploadHandler) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeParamError, "获取文件失败: "+err.Error())
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExts[ext] {
		response.Fail(c, response.CodeParamError, "不支持的文件格式，仅支持 jpg/jpeg/png/gif/webp")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)

	if err := saveFile(file, savePath); err != nil {
		response.Fail(c, response.CodeBizError, "保存文件失败: "+err.Error())
		return
	}

	url := "/uploads/" + filename
	response.OK(c, gin.H{"url": url})
}

func saveFile(src multipart.File, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, src)
	return err
}
