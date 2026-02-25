package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务状态码
const (
	CodeOK           = 200
	CodeParamError   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeBizError     = 500
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeOK,
		Msg:  "success",
		Data: data,
	})
}

// OKWithPage 分页成功响应
func OKWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	OK(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, msg string) {
	httpStatus := http.StatusOK
	switch code {
	case CodeUnauthorized:
		httpStatus = http.StatusUnauthorized
	case CodeForbidden:
		httpStatus = http.StatusForbidden
	}
	c.JSON(httpStatus, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
