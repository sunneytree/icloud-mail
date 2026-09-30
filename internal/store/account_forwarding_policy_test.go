package store_test

import (
	"context"
	"errors"
	"testing"

	"icloud-api/internal/store"
)

func TestAccountForwardingPolicyDefaultsOffAndPersists(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)
	account := createAccount(t, ctx, db, "forwarding", "forwarding@icloud.com")
	if account.VerifyForwardingTarget {
		t.Fatal("new account unexpectedly verifies forwarding target")
	}
	if err := db.SetAccountForwardingPolicy(ctx, account.ID, true); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetAccount(ctx, account.ID)
	if err != nil || !got.VerifyForwardingTarget {
		t.Fatalf("enabled policy: account=%#v err=%v", got, err)
	}
	if _, err := db.UpdateAccount(ctx, got); err != nil {
		t.Fatal(err)
	}
	got, err = db.GetAccount(ctx, account.ID)
	if err != nil || !got.VerifyForwardingTarget {
		t.Fatalf("unrelated update lost policy: account=%#v err=%v", got, err)
	}
	if err := db.SetAccountForwardingPolicy(ctx, account.ID, false); err != nil {
		t.Fatal(err)
	}
	got, err = db.GetAccount(ctx, account.ID)
	if err != nil || got.VerifyForwardingTarget {
		t.Fatalf("disabled policy: account=%#v err=%v", got, err)
	}
	if err := db.SetAccountForwardingPolicy(ctx, account.ID+1000, true); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing account error = %v", err)
	}
}
