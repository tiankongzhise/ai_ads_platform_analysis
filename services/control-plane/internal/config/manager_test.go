package config

import "testing"

type testLoader map[string]string

func (t testLoader) Load() (map[string]string, error) {
	out := map[string]string{}
	for key, value := range t {
		out[key] = value
	}
	return out, nil
}

func TestAuthDefaults(t *testing.T) {
	manager, err := NewManager(testLoader{})
	if err != nil {
		t.Fatal(err)
	}
	auth, err := manager.Auth()
	if err != nil {
		t.Fatal(err)
	}
	if auth.AccessTokenTTL.Hours() != 24 {
		t.Fatalf("expected 24h access token ttl, got %s", auth.AccessTokenTTL)
	}
	if auth.RefreshTokenTTL.Hours() != 720 {
		t.Fatalf("expected 720h refresh token ttl, got %s", auth.RefreshTokenTTL)
	}
}

func TestUpdateValidatesDuration(t *testing.T) {
	manager, err := NewManager(testLoader{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Update("auth.access_token_ttl", "nope", "tester"); err == nil {
		t.Fatal("expected invalid duration error")
	}
	value, err := manager.Update("auth.access_token_ttl", "48h", "tester")
	if err != nil {
		t.Fatal(err)
	}
	if value.Version != 2 {
		t.Fatalf("expected version 2, got %d", value.Version)
	}
}
