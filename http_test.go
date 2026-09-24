package main

import (
	"net/http"
	"net/url"
	"testing"
)

// 验证全局 transport 的 Proxy 接线：应用内代理池未配置（默认态）时，
// HTTPS 请求回退到环境变量代理。
//
// 这里用 clineEnvProxy 替身注入，而不是 t.Setenv("HTTPS_PROXY", ...)。
// http.ProxyFromEnvironment 的解析结果是进程级 sync.Once 缓存：一旦被先运行的
// 测试在「无环境变量」时触发过，就永久固化为「无代理」，此后 t.Setenv 不再生效。
// 那样这个测试的成败取决于执行顺序（实测：单独跑通过、全量跑失败），断言也从
// 「本项目是否接上了环境变量代理」漂移成「标准库缓存何时被谁触发」。
func TestHTTPTransportUsesHTTPSProxyFromEnvironment(t *testing.T) {
	sentinel := &url.URL{Scheme: "http", Host: "127.0.0.1:8080"}
	oldEnvProxy := clineEnvProxy
	clineEnvProxy = func(*http.Request) (*url.URL, error) { return sentinel, nil }
	t.Cleanup(func() { clineEnvProxy = oldEnvProxy })

	// 应用内代理池留空，确保走环境变量回退分支
	setClineProxyTestConfig(t, nil, "round_robin")

	req, err := http.NewRequest(http.MethodPost, "https://api.workos.com/user_management/authorize/device", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	proxyURL, err := httpTransport.Proxy(req)
	if err != nil {
		t.Fatalf("resolve proxy: %v", err)
	}
	if proxyURL == nil {
		t.Fatal("expected env proxy fallback when the app proxy pool is inactive")
	}
	if proxyURL.String() != sentinel.String() {
		t.Fatalf("proxy URL = %q, want %q", proxyURL, sentinel)
	}
}
