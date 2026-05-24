package auth

import (
	"testing"

	"eduadcrm/services/control-plane/internal/config"
	"eduadcrm/services/control-plane/internal/store"
)

type authTestLoader map[string]string

func (t authTestLoader) Load() (map[string]string, error) {
	out := map[string]string{}
	for key, value := range t {
		out[key] = value
	}
	return out, nil
}

func TestRegisterUsesConfiguredTTL(t *testing.T) {
	manager, err := config.NewManager(authTestLoader{})
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(manager, store.NewMemoryStore())
	resp, err := service.Register(RegisterRequest{
		TenantName:  "星海教育",
		Email:       "admin@example.com",
		DisplayName: "Admin",
		Password:    "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	accessHours := resp.AccessTokenExpiresAt.Sub(resp.User.CreatedAt).Hours()
	if accessHours < 23.9 || accessHours > 24.1 {
		t.Fatalf("expected access token around 24h, got %.2f", accessHours)
	}
	refreshHours := resp.RefreshTokenExpiresAt.Sub(resp.User.CreatedAt).Hours()
	if refreshHours < 719.9 || refreshHours > 720.1 {
		t.Fatalf("expected refresh token around 720h, got %.2f", refreshHours)
	}
}
