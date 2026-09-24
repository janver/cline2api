package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 回归：GET /admin/api/opencode/config 的 proxies 必须返回明文。
// 一旦返回脱敏值，前端会把它填进可编辑 textarea 并原样回传，
// 后端整体覆盖后真实凭据就被 "***" 永久破坏了。
func TestOpenCodeConfigReturnsPlainProxies(t *testing.T) {
	const plain = "http://uuu:tk-abc@111.154.222.218:2260"

	zenConfigMu.Lock()
	old := zenConfig
	zenConfig = defaultZenConfig()
	zenConfig.Proxies = []string{plain}
	zenConfigMu.Unlock()
	t.Cleanup(func() {
		zenConfigMu.Lock()
		zenConfig = old
		zenConfigMu.Unlock()
	})

	recorder := httptest.NewRecorder()
	handleOpenCodeConfig(recorder, httptest.NewRequest(http.MethodGet, "/admin/api/opencode/config", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var resp struct {
		Data struct {
			Proxies []string `json:"proxies"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data.Proxies) != 1 {
		t.Fatalf("proxies = %v, want exactly one entry", resp.Data.Proxies)
	}
	if resp.Data.Proxies[0] != plain {
		t.Fatalf("proxies[0] = %q, want plaintext %q", resp.Data.Proxies[0], plain)
	}
}

// 回归：GET /admin/api/cline-proxy/config 的 proxies 必须返回明文，理由同
// TestOpenCodeConfigReturnsPlainProxies —— 前端把返回值填进可编辑 textarea
// 并原样回传，后端整体覆盖后真实凭据就被 "***" 永久破坏了。
func TestClineProxyConfigReturnsPlainProxies(t *testing.T) {
	const plain = "http://uuu:tk-abc@111.154.222.218:2260"

	clineProxyCfgMu.Lock()
	old := clineProxyCfg
	// 直接注入，避免触发 getClineProxyConfig 的磁盘加载与 setClineProxyConfig 的落盘
	clineProxyCfg = &clineProxyConfigData{Proxies: []string{plain}, ProxyStrategy: "round_robin"}
	clineProxyCfgMu.Unlock()
	t.Cleanup(func() {
		clineProxyCfgMu.Lock()
		clineProxyCfg = old
		clineProxyCfgMu.Unlock()
	})

	recorder := httptest.NewRecorder()
	handleClineProxyConfig(recorder, httptest.NewRequest(http.MethodGet, "/admin/api/cline-proxy/config", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var resp struct {
		Data struct {
			Proxies []string `json:"proxies"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data.Proxies) != 1 {
		t.Fatalf("proxies = %v, want exactly one entry", resp.Data.Proxies)
	}
	if resp.Data.Proxies[0] != plain {
		t.Fatalf("proxies[0] = %q, want plaintext %q", resp.Data.Proxies[0], plain)
	}
}
