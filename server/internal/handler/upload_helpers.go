package handler

import (
	"amiya-eden/global"
	"mime/multipart"

	"go.uber.org/zap"
)

func closeUploadFile(file multipart.File) {
	if err := file.Close(); err != nil {
		global.Logger.Error("关闭文件失败", zap.Error(err))
	}
}
