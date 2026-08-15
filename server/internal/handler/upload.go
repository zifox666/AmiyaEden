package handler

import (
	"amiya-eden/pkg/response"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const (
	uploadDir     = "./uploads"
	maxUploadSize = 5 << 20 // 5 MB
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
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

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeParamError, "获取文件失败: "+err.Error())
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, readErr := file.Read(buffer)
	if readErr != nil && readErr != io.EOF {
		response.Fail(c, response.CodeParamError, "读取图片失败")
		return
	}
	contentType := http.DetectContentType(buffer[:n])
	ext, allowed := allowedImageTypes[contentType]
	if !allowed {
		response.Fail(c, response.CodeParamError, "不支持的文件格式，仅支持 jpg/jpeg/png/gif/webp")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.Fail(c, response.CodeBizError, "重置图片读取位置失败")
		return
	}

	randomName := make([]byte, 16)
	if _, err := rand.Read(randomName); err != nil {
		response.Fail(c, response.CodeBizError, "生成图片文件名失败")
		return
	}
	filename := fmt.Sprintf("%s%s", hex.EncodeToString(randomName), ext)
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
