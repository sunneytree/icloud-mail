package hmesync

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"icloud-api/internal/apple"
	"icloud-api/internal/domain"
)

type smsResendAppleClient struct {
	fakeAppleClient
	requestCalls atomic.Int32
}

func (c *smsResendAppleClient) ListTrustedPhones(context.Context, apple.Session) ([]apple.TrustedPhone, apple.Session, error) {
	return []apple.TrustedPhone{{ID: 7, LastTwoDigits: "89"}}, apple.Session{Region: apple.RegionGlobal}, nil
}

func (c *smsResendAppleClient) ListTrustedPhonesWithStatus(context.Context, apple.Session) (apple.TrustedPhonesResult, apple.Session, error) {
	return apple.TrustedPhonesResult{
		Phones:         []apple.TrustedPhone{{ID: 7, LastTwoDigits: "89"}},
		SMSAlreadySent: true,
	}, apple.Session{Region: apple.RegionGlobal}, nil
}

func (c *smsResendAppleClient) RequestSMSCode(context.Context, apple.Session, int) (apple.Session, error) {
	c.requestCalls.Add(1)
	return apple.Session{Region: apple.RegionGlobal}, nil
}

func (c *smsResendAppleClient) VerifySMSCode(context.Context, apple.Session, int, string) (apple.Session, error) {
	return apple.Session{Region: apple.RegionGlobal}, nil
}

func TestRequestSMSAuthResendsAfterAutomatic201Delivery(t *testing.T) {
	now := time.Now().UTC()
	repo := newFakeRepository(domain.Account{ID: 3, Email: "primary@icloud.com", Enabled: true}, now)
	client := &smsResendAppleClient{}
	client.signIn = func(_ context.Context, appleID, _ string, region apple.Region, _ *apple.Session) (apple.Session, bool, error) {
		return apple.Session{AppleID: appleID, Region: region, SessionToken: "challenge"}, true, nil
	}
	service := newTestService(t, repo, client, &fakeLocker{}, func() time.Time { return now })

	started, err := service.StartAuth(context.Background(), 1, 3, "primary@icloud.com", "password", apple.RegionGlobal)
	if err != nil || started.Status != StatusVerificationRequired {
		t.Fatalf("start auth = %#v, err=%v", started, err)
	}

	first, err := service.RequestSMSAuth(context.Background(), 1, 3, started.ChallengeID, 0)
	if err != nil || first.Status != StatusSMSSent || first.SelectedPhone == nil {
		t.Fatalf("first SMS request = %#v, err=%v", first, err)
	}
	if got := client.requestCalls.Load(); got != 0 {
		t.Fatalf("first HTTP 201 path called PUT %d times, want 0", got)
	}

	second, err := service.RequestSMSAuth(context.Background(), 1, 3, started.ChallengeID, first.SelectedPhone.ID)
	if err != nil || second.Status != StatusSMSSent || second.SelectedPhone == nil {
		t.Fatalf("resend SMS request = %#v, err=%v", second, err)
	}
	if got := client.requestCalls.Load(); got != 1 {
		t.Fatalf("resend path called PUT %d times, want 1", got)
	}
}
