package httpserver

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"icloud-api/internal/apple"
	"icloud-api/internal/applog"
	"icloud-api/internal/hmesync"
)

func TestAppleSessionDiagnosticsKeepUpstreamStageWithoutResponseData(t *testing.T) {
	upstream := &apple.Error{Op: "validate Apple session", Kind: apple.ErrInvalidSession, StatusCode: 421, ServiceCode: "FIXTURE_CODE"}
	err := errors.Join(hmesync.ErrSessionExpired, upstream)
	result := classifyAdminAPIAppleError(err)
	if result.Code != hmesync.CodeSessionExpired || result.UpstreamOperation != upstream.Op || result.UpstreamStatus != 421 || result.UpstreamServiceCode != "FIXTURE_CODE" {
		t.Fatalf("lost session diagnostics: %#v", result)
	}
	if strings.Contains(result.Message, "FIXTURE_CODE") {
		t.Fatal("raw diagnostic leaked into user message")
	}
	if !strings.Contains(result.Message, "上游操作：validate Apple session") || !strings.Contains(result.Message, "上游响应：HTTP 421") {
		t.Fatalf("upstream detail missing from user message: %s", result.Message)
	}
}

func TestAppleUpstream423ExplainsLockedChallenge(t *testing.T) {
	upstream := &apple.Error{Op: "list trusted phone numbers", Kind: apple.ErrService, StatusCode: 423}
	result := classifyAdminAPIAppleError(errors.Join(hmesync.ErrUpstream, upstream))
	if result.Code != hmesync.CodeUpstreamError {
		t.Fatalf("code = %q, want %q", result.Code, hmesync.CodeUpstreamError)
	}
	want := "Apple 服务暂时异常，请稍后再试（上游操作：读取受信任手机号；上游响应：HTTP 423（Apple 当前验证挑战可能已锁定或暂不可用））"
	if result.Message != want {
		t.Fatalf("message = %q, want %q", result.Message, want)
	}
}

func TestWebSessionDiagnosticsAreReadableAndNonSensitiveInAppLog(t *testing.T) {
	handler := applog.New(10)
	slog.New(handler).Warn("Apple 管理操作失败", "hme_request_diagnostics", apple.WebSessionDiagnostics{
		Region: apple.RegionChina, ServiceRegion: apple.RegionGlobal, MatchingCookies: 2,
		WebAuthPresent: true, WebUserPresent: false, HMEActive: true, HMEAvailable: true,
	})
	items := handler.List(applog.Filter{Limit: 10}).Items
	if len(items) != 1 {
		t.Fatalf("log item count = %d, want 1", len(items))
	}
	fields := items[0].Fields
	if fields["hme_request_diagnostics.region"] != string(apple.RegionChina) ||
		fields["hme_request_diagnostics.service_region"] != string(apple.RegionGlobal) ||
		fields["hme_request_diagnostics.matching_items"] != "2" ||
		fields["hme_request_diagnostics.web_auth_present"] != "true" {
		t.Fatalf("diagnostics were not readable: %#v", fields)
	}
	for key, value := range fields {
		if strings.Contains(strings.ToLower(value), "token") || strings.Contains(strings.ToLower(value), "cookie") || strings.Contains(strings.ToLower(value), "id=") {
			t.Fatalf("sensitive diagnostic leaked: %q=%q", key, value)
		}
	}
}

func TestDirectoryAuthenticationDiagnosticsAreDistinctFromExpiredLogin(t *testing.T) {
	diagnostics := &apple.WebSessionDiagnostics{Region: apple.RegionChina, ServiceRegion: apple.RegionGlobal, MatchingCookies: 0}
	upstream := &apple.Error{Op: "list Hide My Email aliases", Kind: apple.ErrHMEAuthentication, StatusCode: 401, WebSession: diagnostics}
	result := classifyAdminAPIAppleError(upstream)
	if result.Code != hmesync.CodeHMEAuthentication || result.UpstreamStatus != 401 || result.WebSessionDiagnostics != diagnostics {
		t.Fatalf("lost directory diagnostics: %#v", result)
	}
	if !strings.Contains(result.Message, "HTTP 401") || !strings.Contains(result.Message, "登录状态已保留") || strings.Contains(result.Message, "会话已过期") {
		t.Fatalf("misleading directory error: %s", result.Message)
	}
}
