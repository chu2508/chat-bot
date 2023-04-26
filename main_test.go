package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"tastien.com/chat-bot/bot"
)

func TestPingRoute(t *testing.T) {
	cfg := loadTestConfig(t)

	_, r := CreateApp(cfg)

	// 创建一个模拟的 HTTP 请求
	req, err := http.NewRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 使用 httptest 包创建一个 ResponseRecorder 来记录响应
	w := httptest.NewRecorder()

	// 使用 Gin 的 ServeHTTP 方法来处理请求和记录响应
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应体
	assert.Equal(t, "{\"message\":\"pong\"}", w.Body.String())
}
func TestWebhookRoute(t *testing.T) {

	// 加载测试数据文件
	file, err := os.Open("message.example.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	cfg := loadTestConfig(t)

	_, r := CreateApp(cfg)

	// 创建一个模拟的 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, "/webhook/event", file)
	if err != nil {
		t.Fatal(err)
	}

	// 使用 httptest 包创建一个 ResponseRecorder 来记录响应
	w := httptest.NewRecorder()

	// 使用 Gin 的 ServeHTTP 方法来处理请求和记录响应
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应体
	assert.Equal(t, "{\"msg\":\"success\"}", w.Body.String())
}
func TestWebhookRouteChallenge(t *testing.T) {

	cfg := loadTestConfig(t)

	_, r := CreateApp(cfg)

	body := map[string]string{
		"challenge": "ajls384kdjx98XX",              // 应用需要在响应中原样返回的值
		"token":     cfg.FeishuAppVerificationToken, // 即 Verification Token
		"type":      "url_verification",             // 表示这是一个验证请求
	}
	data, _ := json.Marshal(body)

	// 创建一个模拟的 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, "/webhook/event", strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	// 使用 httptest 包创建一个 ResponseRecorder 来记录响应
	w := httptest.NewRecorder()

	// 使用 Gin 的 ServeHTTP 方法来处理请求和记录响应
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应体
	assert.Equal(t, "{\"challenge\":\"ajls384kdjx98XX\"}", w.Body.String())
}

func loadTestConfig(t *testing.T) *bot.Config {
	cfg := &bot.Config{}
	cfgFile, err := os.Open("config.test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer cfgFile.Close()

	if data, err := ioutil.ReadAll(cfgFile); err == nil {
		json.Unmarshal(data, cfg)
	}
	return cfg
}

func TestWebHookAddUser(t *testing.T) {

	// 加载测试数据文件
	file, err := os.Open("userAdded.example.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	cfg := loadTestConfig(t)

	_, r := CreateApp(cfg)

	// 创建一个模拟的 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, "/webhook/event", file)
	if err != nil {
		t.Fatal(err)
	}

	// 使用 httptest 包创建一个 ResponseRecorder 来记录响应
	w := httptest.NewRecorder()

	// 使用 Gin 的 ServeHTTP 方法来处理请求和记录响应
	r.ServeHTTP(w, req)

	// 检查响应状态码
	assert.Equal(t, http.StatusOK, w.Code)

	// 检查响应体
	assert.Equal(t, "{\"msg\":\"success\"}", w.Body.String())
}
