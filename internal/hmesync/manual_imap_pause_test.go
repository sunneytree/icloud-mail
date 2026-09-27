package hmesync

import (
	"context"
	"errors"
	"testing"
	"time"

	"icloud-api/internal/apple"
	"icloud-api/internal/domain"
)

func TestManualProbeIgnoresPausedIMAPAuthentication(t *testing.T) {
	now := time.Now().UTC()
	repo := newFakeRepository(domain.Account{
		ID: 1, Email: "owner@icloud.com", Enabled: true,
		LastSyncError: "login IMAP account: imap: NO [AUTHENTICATIONFAILED] Authentication Failed",
	}, now)
	validations := 0
	client := &fakeAppleClient{validate: func(_ context.Context, session apple.Session) (apple.Session, error) {
		validations++
		return session, errors.New("validation reached")
	}}
	service := newTestService(t, repo, client, &fakeLocker{}, func() time.Time { return now })
	storeSession(t, service, repo, 1, apple.Session{
		AppleID: "owner@icloud.com", Region: apple.RegionGlobal,
		SessionToken: "fixture", DSID: "42",
	})
	_, err := service.ProbeAliasWithChannel(context.Background(), 1, "auto")
	if validations != 1 || errors.Is(err, domain.ErrIMAPAuthenticationPaused) {
		t.Fatalf("manual probe validations=%d error=%v; want Apple validation without IMAP pause", validations, err)
	}
}
