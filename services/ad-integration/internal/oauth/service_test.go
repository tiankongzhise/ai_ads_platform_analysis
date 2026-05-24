package oauth

import (
	"net/url"
	"testing"

	"eduadcrm/services/ad-integration/internal/store"
)

func TestStateCanOnlyBeConsumedOnce(t *testing.T) {
	service := NewService(store.NewMemoryStore())
	auth, err := service.Authorize("douyin", AuthorizeRequest{
		TenantID:       "tenant-1",
		UserID:         "user-1",
		OrganizationID: "org-1",
		ChannelID:      "channel-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	query := url.Values{}
	query.Set("state", auth.State)
	query.Set("auth_code", "code-1")
	if _, err := service.Callback("douyin", query); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Callback("douyin", query); err == nil {
		t.Fatal("expected replayed state to fail")
	}
}
