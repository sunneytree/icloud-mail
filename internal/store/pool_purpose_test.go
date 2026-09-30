package store_test

import (
	"context"
	"errors"
	"fmt"
	"icloud-api/internal/store"
	"os"
	"sync"
	"testing"
)

func TestPoolPurposeIsolation(t *testing.T) { runPoolPurposeIsolation(t, "") }
func TestPoolPurposePostgres(t *testing.T) {
	dsn := os.Getenv("ICLOUD_POOL_TEST_DSN")
	if dsn == "" {
		t.Skip("requires dedicated test database")
	}
	runPoolPurposeIsolation(t, dsn)
}
func runPoolPurposeIsolation(t *testing.T, dsn string) {
	ctx := context.Background()
	db, _, _, aliases, client := poolFixture(t, 1, dsn)
	claim := func(id, purpose string) ([]store.PoolLease, error) {
		return db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: id, Purpose: purpose, Count: 1, TTLSeconds: 600, ImmediateUse: true})
	}
	adobe, e := claim("purpose-adobe-01", " Adobe ")
	if e != nil {
		t.Fatal(e)
	}
	if adobe[0].Purpose != "adobe" {
		t.Fatal(adobe)
	}
	before, _ := db.GetAlias(ctx, aliases[0].ID)
	gpt, e := claim("purpose-gpt-0001", "gpt")
	if e != nil {
		t.Fatal(e)
	}
	after, _ := db.GetAlias(ctx, aliases[0].ID)
	if before.CredentialVersion != after.CredentialVersion {
		t.Fatal("cross-purpose claim rotated shared credentials")
	}
	if _, e = claim("purpose-adobe-02", "adobe"); !errors.Is(e, store.ErrPoolEmpty) {
		t.Fatal(e)
	}
	retry, e := claim("purpose-adobe-01", "adobe")
	if e != nil || retry[0].ID != adobe[0].ID {
		t.Fatal(retry, e)
	}
	if _, e = claim("purpose-adobe-01", "gpt"); !errors.Is(e, store.ErrPoolRequestConflict) {
		t.Fatal(e)
	}
	common, e := claim("purpose-common01", "")
	if e != nil || common[0].Purpose != "common" {
		t.Fatal(common, e)
	}
	if _, e = claim("purpose-common02", "  "); !errors.Is(e, store.ErrPoolEmpty) {
		t.Fatal(e)
	}
	if _, e = claim("purpose-invalid1", "bad purpose"); !errors.Is(e, store.ErrPoolInput) {
		t.Fatal(e)
	}
	for _, purpose := range []string{"adobe", "gpt", "common"} {
		page, e := db.ListPoolMembers(ctx, "used", "", 50, 0, purpose)
		if e != nil || page.Total != 1 {
			t.Fatal(purpose, page, e)
		}
	}
	if _, e = db.ActPoolLease(ctx, adobe[0].ID, client.ID, "release_used", 0); e != nil {
		t.Fatal(e)
	}
	after, _ = db.GetAlias(ctx, aliases[0].ID)
	if before.CredentialVersion != after.CredentialVersion {
		t.Fatal("release invalidated another purpose")
	}
	g, e := db.GetPoolLease(ctx, gpt[0].ID, client.ID)
	if e != nil || g.State != "used" {
		t.Fatal(g, e)
	}
	if _, e = claim("purpose-adobe-03", "adobe"); e != nil {
		t.Fatal(e)
	}
	page, e := db.ListPoolMembers(ctx, "available", "", 50, 0, "fresh")
	if e != nil || page.Total != 1 {
		t.Fatal(page, e)
	}
}
func TestPoolPurposeConcurrent(t *testing.T) {
	ctx := context.Background()
	db, _, _, _, client := poolFixture(t, 1, "")
	var wg sync.WaitGroup
	var mu sync.Mutex
	counts := map[string]int{}
	for _, p := range []string{"adobe", "gpt"} {
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func(p string, i int) {
				defer wg.Done()
				_, e := db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: fmt.Sprintf("purpose-%s-%03d", p, i), Purpose: p, Count: 1, TTLSeconds: 600, ImmediateUse: true})
				mu.Lock()
				defer mu.Unlock()
				if e == nil {
					counts[p]++
				} else if !errors.Is(e, store.ErrPoolEmpty) {
					t.Error(e)
				}
			}(p, i)
		}
	}
	wg.Wait()
	if counts["adobe"] != 1 || counts["gpt"] != 1 {
		t.Fatal(counts)
	}
}
func TestPoolPurposeMigratesOnce(t *testing.T) {
	ctx := context.Background()
	db, _, _, _, client := poolFixture(t, 2, "")
	used, e := db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: "legacy-used-0001", Count: 1, TTLSeconds: 600, ImmediateUse: true})
	if e != nil {
		t.Fatal(e)
	}
	// Simulate a pre-purpose database, preserving used claims and legacy fingerprints.
	for _, q := range []string{`DROP INDEX pool_one_live_purpose_lease`, `ALTER TABLE pool_leases DROP COLUMN purpose`, `UPDATE pool_requests SET fingerprint='{"request_id":"legacy-used-0001","count":1,"account_id":0,"group_id":0,"ttl_seconds":600}'`} {
		if _, e = db.DB().Exec(q); e != nil {
			t.Fatal(e)
		}
	}
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	got, e := db.GetPoolLease(ctx, used[0].ID, client.ID)
	if e != nil || got.Purpose != "adobe" {
		t.Fatal(got, e)
	}
	retry, e := db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: "legacy-used-0001", Purpose: "adobe", Count: 1, TTLSeconds: 600, ImmediateUse: true})
	if e != nil || retry[0].ID != used[0].ID {
		t.Fatal(retry, e)
	}
	common, e := db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: "new-common-0001", Count: 1, TTLSeconds: 600, ImmediateUse: true})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.Migrate(ctx); e != nil {
		t.Fatal(e)
	}
	got, e = db.GetPoolLease(ctx, common[0].ID, client.ID)
	if e != nil || got.Purpose != "common" {
		t.Fatal(got, e)
	}
}

func TestPoolPurposeLegacyMarkKeepsCredentials(t *testing.T) {
	ctx := context.Background()
	db, _, _, aliases, client := poolFixture(t, 1, "")
	if _, err := db.DB().Exec(`INSERT INTO pool_legacy_purpose_marks(alias_id,purpose) VALUES(?,'adobe')`, aliases[0].ID); err != nil {
		t.Fatal(err)
	}
	before, _ := db.GetAlias(ctx, aliases[0].ID)
	leases, err := db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: "legacy-gpt-0001", Purpose: "gpt", Count: 1, TTLSeconds: 600, ImmediateUse: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ActPoolLease(ctx, leases[0].ID, client.ID, "release_used", 0); err != nil {
		t.Fatal(err)
	}
	after, _ := db.GetAlias(ctx, aliases[0].ID)
	if before.CredentialVersion != after.CredentialVersion {
		t.Fatal("legacy mark credentials rotated")
	}
	if err = db.SetPoolMember(ctx, aliases[0].ID, "paused"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ClaimPool(ctx, client.ID, store.PoolClaim{RequestID: "paused-gpt-0001", Purpose: "gpt", Count: 1, TTLSeconds: 600}); !errors.Is(err, store.ErrPoolEmpty) {
		t.Fatal(err)
	}
	if err = db.SetPoolMember(ctx, aliases[0].ID, "available"); err != nil {
		t.Fatal(err)
	}
	page, err := db.ListPoolMembers(ctx, "used", "", 50, 0, "adobe")
	if err != nil || page.Total != 1 {
		t.Fatal(page, err)
	}
}
