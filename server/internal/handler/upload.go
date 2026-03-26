package handler

import (
	"amiya-eden/pkg/response"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	maxUploadSize = 2048 << 10 // 2MB
)

var uploadAllowedMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadHandler 文件上传处理器
type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

// UploadImage POST /api/v1/upload/image
// 上传图片，返回 base64 data URL（不写入文件系统）
func (h *UploadHandler) UploadImage(c *gin.Context) {
	uploadImageAsDataURL(c, maxUploadSize, uploadAllowedMIME)
}

func uploadImageAsDataURL(c *gin.Context, maxBytes int64, allowedMIME map[string]bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes+1024)

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, response.CodeParamError, "获取文件失败: "+err.Error())
		return
	}
	defer closeUploadFile(file)

	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		response.Fail(c, response.CodeBizError, "读取文件失败")
		return
	}
	if int64(len(data)) > maxBytes {
		response.Fail(c, response.CodeParamError, fmt.Sprintf("图片大小不能超过 %d KB", maxBytes>>10))
		return
	}

	mime := http.DetectContentType(data)
	if !allowedMIME[mime] {
		response.Fail(c, response.CodeParamError, "仅支持 jpeg/png/webp 格式")
		return
	}

	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))
	response.OK(c, gin.H{"url": dataURL})
}
