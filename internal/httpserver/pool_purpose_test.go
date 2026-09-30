package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"icloud-api/internal/store"
	"testing"
)

func TestPoolPurposeAPI(t *testing.T) {
	env := newAdminAPITestEnv(t)
	env.server.cfg.PoolClaimImmediateUsed = true
	ctx := context.Background()
	account := adminAPITestCreateAccount(t, env, "purpose-owner@icloud.com")
	alias, _ := createV2AliasFixture(t, env, account.ID, "purpose-alias@icloud.com")
	if e := env.store.EnrollPoolAliases(ctx, []int64{alias.ID}); e != nil {
		t.Fatal(e)
	}
	_, key, e := env.store.CreatePoolClient(ctx, "purpose-api")
	if e != nil {
		t.Fatal(e)
	}
	router, e := env.server.Router()
	if e != nil {
		t.Fatal(e)
	}
	headers := map[string]string{"Authorization": "Bearer " + key, "Content-Type": "application/json"}
	ids := map[string]string{}
	for i, purpose := range []string{"adobe", "gpt", "", "adobe"} {
		body := fmt.Sprintf(`{"request_id":"purpose-api-%04d","count":1,"purpose":%q}`, i, purpose)
		res := serveV2Request(router, "POST", "/api/v1/pool/claim", body, headers)
		want := 200
		if i == 3 {
			want = 409
		}
		if res.Code != want {
			t.Fatalf("%s got %d %s", purpose, res.Code, res.Body.String())
		}
		if want != 200 {
			continue
		}
		var out struct {
			Data struct {
				Leases []struct {
					Lease store.PoolLease `json:"lease"`
				} `json:"leases"`
			} `json:"data"`
		}
		if e := json.Unmarshal(res.Body.Bytes(), &out); e != nil {
			t.Fatal(e)
		}
		expected := purpose
		if expected == "" {
			expected = "common"
		}
		if out.Data.Leases[0].Lease.Purpose != expected {
			t.Fatal(out)
		}
		ids[expected] = out.Data.Leases[0].Lease.ID
	}
	release := serveV2Request(router, "POST", "/api/v1/pool/leases/"+ids["adobe"]+"/release", "", headers)
	if release.Code != 200 {
		t.Fatal(release.Body.String())
	}
	list := serveV2Request(router, "GET", "/api/v1/pool/leases?purpose=gpt", "", headers)
	if list.Code != 200 {
		t.Fatal(list.Body.String())
	}
	var out struct {
		Data store.PoolPage[store.PoolLease] `json:"data"`
	}
	json.Unmarshal(list.Body.Bytes(), &out)
	if out.Data.Total != 1 || out.Data.Items[0].State != "used" || out.Data.Items[0].Purpose != "gpt" {
		t.Fatal(out)
	}
}
