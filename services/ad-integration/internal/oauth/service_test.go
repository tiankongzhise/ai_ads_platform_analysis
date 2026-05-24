package oauth

import (
	"net/url"
	"testing"
	"time"

	"eduadcrm/services/ad-integration/internal/store"
)

type fakeQueue struct {
	queueName string
	message   store.QueueMessage
	called    bool
}

func (q *fakeQueue) Publish(queueName string, message store.QueueMessage) error {
	q.queueName = queueName
	q.message = message
	q.called = true
	return nil
}

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

func TestCallbackPublishesAdSyncMessage(t *testing.T) {
	repository := store.NewMemoryStore()
	queue := &fakeQueue{}
	service := oauthWithQueue(repository, queue)
	auth, err := service.Authorize("tencent", AuthorizeRequest{
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
	query.Set("authorization_code", "code-1")
	query.Set("account_id", "account-1")
	if _, err := service.Callback("tencent", query); err != nil {
		t.Fatal(err)
	}
	if !queue.called {
		t.Fatal("expected queue publish")
	}
	if queue.queueName != "ad_sync_queue" {
		t.Fatalf("expected ad_sync_queue, got %s", queue.queueName)
	}
	if queue.message.EventType != "ad.sync.requested" {
		t.Fatalf("expected ad.sync.requested, got %s", queue.message.EventType)
	}
	if queue.message.OccurredAt.After(time.Now().UTC().Add(time.Second)) {
		t.Fatalf("unexpected occurred_at: %s", queue.message.OccurredAt)
	}
}

func oauthWithQueue(repository store.Repository, queue store.MessageQueue) *Service {
	return NewServiceWithQueue(repository, queue)
}
