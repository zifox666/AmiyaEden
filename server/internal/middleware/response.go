package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

// bodyWriter 拦截 gin 响应体，写入本地缓冲而不直接发送
type bodyWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.body.Write(b)
}

func (w *bodyWriter) WriteString(s string) (int, error) {
	w.written = true
	return w.body.WriteString(s)
}

func (w *bodyWriter) WriteHeader(code int) {
	w.statusCode = code
}

// WriteHeaderNow noop：阻止 gin 内部提前发送响应头
func (w *bodyWriter) WriteHeaderNow() {}

func (w *bodyWriter) Written() bool { return w.written }

func (w *bodyWriter) Status() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *bodyWriter) Size() int { return w.body.Len() }

// ResponseWrapper 统一包装响应体为 {code, msg, data} 格式。
// 若响应体已包含 "code" 字段（即 handler 已调用 response.OK/Fail），则直接透传；
// 若为其他 JSON，则自动包装；非 JSON（文件下载等）直接透传。
func ResponseWrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		bw := &bodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = bw

		c.Next()

		// 恢复原始 writer，准备写最终响应
		c.Writer = bw.ResponseWriter

		statusCode := bw.Status()
		body := bw.body.Bytes()

		// ---- 无响应体 ----
		if len(body) == 0 {
			// 对 404/405 等框架级错误包装为标准格式
			if statusCode == http.StatusNotFound {
				writeJSON(c, statusCode, response.Response{
					Code: response.CodeNotFound,
					Msg:  "资源不存在",
					Data: nil,
				})
				return
			}
			if statusCode == http.StatusMethodNotAllowed {
				writeJSON(c, statusCode, response.Response{
					Code: response.CodeBizError,
					Msg:  "方法不允许",
					Data: nil,
				})
				return
			}
			c.Writer.WriteHeader(statusCode)
			return
		}

		// ---- 非 JSON 响应（文件下载等）透传 ----
		ct := bw.Header().Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			c.Writer.WriteHeader(statusCode)
			_, _ = c.Writer.Write(body)
			return
		}

		// ---- 检查是否已经是标准格式 ----
		var check map[string]json.RawMessage
		if err := json.Unmarshal(body, &check); err == nil {
			if rawCode, ok := check["code"]; ok {
				// 提取 biz_code 写入 context，供 OperationLog 读取
				var bizCode int
				if jerr := json.Unmarshal(rawCode, &bizCode); jerr == nil {
					c.Set(CtxKeyBizCode, bizCode)
				}
				// 已包含 code 字段，直接透传
				c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
				c.Writer.WriteHeader(statusCode)
				_, _ = c.Writer.Write(body)
				return
			}
		}

		// ---- 自动包装裸 JSON ----
		c.Set(CtxKeyBizCode, response.CodeOK)
		wrapped, _ := json.Marshal(response.Response{
			Code: response.CodeOK,
			Msg:  "success",
			Data: json.RawMessage(body),
		})
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.WriteHeader(statusCode)
		_, _ = c.Writer.Write(wrapped)
	}
}

func writeJSON(c *gin.Context, statusCode int, v any) {
	b, _ := json.Marshal(v)
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(statusCode)
	_, _ = c.Writer.Write(b)
}
