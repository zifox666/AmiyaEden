package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"amiya-eden/pkg/response"

	"github.com/gin-gonic/gin"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aF9sAAAAASUVORK5CYII="

func TestUploadImageReturnsDataURL(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	fileBytes, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	req := newUploadImageRequest(t, "tiny.png", fileBytes)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	handler := NewUploadHandler()
	handler.UploadImage(ctx)

	var resp response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != response.CodeOK {
		t.Fatalf("expected success code, got %d with body %s", resp.Code, recorder.Body.String())
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %#v", resp.Data)
	}
	url, _ := data["url"].(string)
	if !strings.HasPrefix(url, "data:image/png;base64,") {
		t.Fatalf("expected base64 data url, got %q", url)
	}
}

func TestUploadImageRejectsFilesOverTwoMB(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	fileBytes := bytes.Repeat([]byte("a"), (2048<<10)+1)
	req := newUploadImageRequest(t, "too-large.png", fileBytes)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	handler := NewUploadHandler()
	handler.UploadImage(ctx)

	var resp response.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != response.CodeParamError {
		t.Fatalf("expected param error, got %d with body %s", resp.Code, recorder.Body.String())
	}
	if !strings.Contains(resp.Msg, "2048 KB") {
		t.Fatalf("expected 2MB validation message, got %q", resp.Msg)
	}
}

func newUploadImageRequest(t *testing.T, filename string, fileBytes []byte) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload/image", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
