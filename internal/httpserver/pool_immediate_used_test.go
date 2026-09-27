package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"icloud-api/internal/store"
)

func TestPoolClaimImmediateUsedAndExplicitRelease(t *testing.T) {
	env := newAdminAPITestEnv(t)
	env.server.cfg.PoolClaimImmediateUsed = true
	ctx := context.Background()
	account := adminAPITestCreateAccount(t, env, "immediate-used@icloud.com")
	alias, _ := createV2AliasFixture(t, env, account.ID, "immediate-alias@icloud.com")
	if err := env.store.EnrollPoolAliases(ctx, []int64{alias.ID}); err != nil {
		t.Fatal(err)
	}
	_, key, err := env.store.CreatePoolClient(ctx, "immediate-project")
	if err != nil {
		t.Fatal(err)
	}
	router, err := env.server.Router()
	if err != nil {
		t.Fatal(err)
	}
	headers := map[string]string{"Authorization": "Bearer " + key, "Content-Type": "application/json"}
	claim := serveV2Request(router, "POST", "/api/v1/pool/claim", `{"request_id":"immediate-used-0001","count":1,"ttl_seconds":60}`, headers)
	if claim.Code != 200 {
		t.Fatalf("claim: %d %s", claim.Code, claim.Body.String())
	}
	var payload struct {
		Data struct {
			Leases []struct {
				Lease   store.PoolLease `json:"lease"`
				Mailbox struct {
					APIKey string `json:"api_key"`
				} `json:"mailbox"`
			} `json:"leases"`
		} `json:"data"`
	}
	if err := json.Unmarshal(claim.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data.Leases) != 1 || payload.Data.Leases[0].Lease.State != "used" || payload.Data.Leases[0].Mailbox.APIKey == "" {
		t.Fatalf("unexpected claim: %s", claim.Body.String())
	}
	lease := payload.Data.Leases[0].Lease
	var memberState string
	if err := env.store.DB().QueryRowContext(ctx, `SELECT state FROM pool_members WHERE alias_id=?`, alias.ID).Scan(&memberState); err != nil || memberState != "used" {
		t.Fatalf("member state = %q, %v", memberState, err)
	}
	if _, err := env.store.DB().ExecContext(ctx, `UPDATE pool_leases SET expires_at=? WHERE id=?`, time.Now().Add(-time.Hour).UnixNano(), lease.ID); err != nil {
		t.Fatal(err)
	}
	if err := env.store.RefreshPool(ctx); err != nil {
		t.Fatal(err)
	}
	stillUsed, err := env.store.GetPoolLease(ctx, lease.ID, "")
	if err != nil || stillUsed.State != "used" {
		t.Fatalf("used lease expired: %+v %v", stillUsed, err)
	}
	unavailable := serveV2Request(router, "POST", "/api/v1/pool/claim", `{"request_id":"immediate-used-0002","count":1,"ttl_seconds":60}`, headers)
	if unavailable.Code != 409 {
		t.Fatalf("used alias reissued: %d %s", unavailable.Code, unavailable.Body.String())
	}
	released := serveV2Request(router, "POST", "/api/v1/pool/leases/"+lease.ID+"/release", "", headers)
	if released.Code != 200 {
		t.Fatalf("release: %d %s", released.Code, released.Body.String())
	}
	if stale := serveV2Request(router, "GET", "/api/v1/otp", "", map[string]string{"Authorization": "Bearer " + payload.Data.Leases[0].Mailbox.APIKey}); stale.Code != 401 {
		t.Fatalf("released credential still works: %d", stale.Code)
	}
	next := serveV2Request(router, "POST", "/api/v1/pool/claim", `{"request_id":"immediate-used-0003","count":1,"ttl_seconds":60}`, headers)
	if next.Code != 200 {
		t.Fatalf("reclaim: %d %s", next.Code, next.Body.String())
	}
}

func TestPoolSettingsSwitchPreservesLeasePolicy(t *testing.T) {
	env := newAdminAPITestEnv(t)
	env.server.cfg.PoolClaimImmediateUsed = true
	ctx := context.Background()
	account := adminAPITestCreateAccount(t, env, "policy@icloud.com")
	alias, _ := createV2AliasFixture(t, env, account.ID, "policy-alias@icloud.com")
	if err := env.store.EnrollPoolAliases(ctx, []int64{alias.ID}); err != nil {
		t.Fatal(err)
	}
	_, key, err := env.store.CreatePoolClient(ctx, "policy-project")
	if err != nil {
		t.Fatal(err)
	}
	cookie, csrf, _ := env.createSession(t, "policy-admin", "test-password")
	get := env.request(t, "GET", "/admin/api/v1/pool/settings", nil, "", []*http.Cookie{cookie}, "")
	if get.Code != 200 || !strings.Contains(get.Body.String(), `"immediate_used":true`) {
		t.Fatalf("default settings: %d %s", get.Code, get.Body.String())
	}
	put := env.request(t, "PUT", "/admin/api/v1/pool/settings", []byte(`{"immediate_used":false}`), "application/json", []*http.Cookie{cookie}, csrf)
	if put.Code != 200 {
		t.Fatalf("set false: %d %s", put.Code, put.Body.String())
	}
	router, err := env.server.Router()
	if err != nil {
		t.Fatal(err)
	}
	headers := map[string]string{"Authorization": "Bearer " + key, "Content-Type": "application/json"}
	claim := serveV2Request(router, "POST", "/api/v1/pool/claim", `{"request_id":"switch-policy-0001","count":1}`, headers)
	if claim.Code != 200 {
		t.Fatalf("claim: %d %s", claim.Code, claim.Body.String())
	}
	var result struct {
		Data struct {
			Leases []struct {
				Lease store.PoolLease `json:"lease"`
			} `json:"leases"`
		} `json:"data"`
	}
	if err := json.Unmarshal(claim.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	lease := result.Data.Leases[0].Lease
	if lease.State != "leased" || lease.ImmediateUse {
		t.Fatalf("wrong lease policy: %+v", lease)
	}
	commit := serveV2Request(router, "POST", "/api/v1/pool/leases/"+lease.ID+"/commit", "", headers)
	if commit.Code != 200 {
		t.Fatalf("commit: %d %s", commit.Code, commit.Body.String())
	}
	put = env.request(t, "PUT", "/admin/api/v1/pool/settings", []byte(`{"immediate_used":true}`), "application/json", []*http.Cookie{cookie}, csrf)
	if put.Code != 200 {
		t.Fatalf("set true: %d %s", put.Code, put.Body.String())
	}
	release := serveV2Request(router, "POST", "/api/v1/pool/leases/"+lease.ID+"/release", "", headers)
	if release.Code != 409 {
		t.Fatalf("old committed lease released: %d %s", release.Code, release.Body.String())
	}
}
