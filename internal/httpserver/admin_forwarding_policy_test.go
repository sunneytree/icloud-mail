package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestAdminForwardingPolicyDefaultsOffAndToggles(t *testing.T) {
	env := newAdminAPITestEnv(t)
	cookie, csrf, _ := env.createSession(t, "forwarding-policy-admin", "password")
	account := adminAPITestCreateAccount(t, env, "forwarding-policy@icloud.com")
	path := fmt.Sprintf("/admin/api/v1/accounts/%d/forwarding-policy", account.ID)
	if account.VerifyForwardingTarget {
		t.Fatal("new account policy should default off")
	}
	unauth := env.request(t, http.MethodPut, path, adminAPITestJSON(t, map[string]bool{"verify_forwarding_target": true}), "application/json", nil, csrf)
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d", unauth.Code)
	}
	withoutCSRF := env.request(t, http.MethodPut, path, adminAPITestJSON(t, map[string]bool{"verify_forwarding_target": true}), "application/json", []*http.Cookie{cookie}, "")
	if withoutCSRF.Code != http.StatusForbidden {
		t.Fatalf("without CSRF status=%d", withoutCSRF.Code)
	}
	for _, enabled := range []bool{true, false} {
		response := env.request(t, http.MethodPut, path, adminAPITestJSON(t, map[string]bool{"verify_forwarding_target": enabled}), "application/json", []*http.Cookie{cookie}, csrf)
		if response.Code != http.StatusOK {
			t.Fatalf("toggle %v status=%d body=%s", enabled, response.Code, response.Body.String())
		}
		var result struct {
			Data struct {
				Verify bool `json:"verify_forwarding_target"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || result.Data.Verify != enabled {
			t.Fatalf("toggle %v response=%s err=%v", enabled, response.Body.String(), err)
		}
		stored, err := env.store.GetAccount(context.Background(), account.ID)
		if err != nil || stored.VerifyForwardingTarget != enabled {
			t.Fatalf("toggle %v stored=%#v err=%v", enabled, stored, err)
		}
	}
}
