package model

import "testing"

func TestDefaultConfigDoesNotShipToken(t *testing.T) {
	if token := NewDefaultFileConfig().Token; token != "" {
		t.Fatalf("default token must be empty, got %q", token)
	}
}

func TestClearLegacyDefaultTokenKeepsPersonalCredential(t *testing.T) {
	config := &FileConfig{Token: "personal-token"}
	if config.ClearLegacyDefaultToken() || config.Token != "personal-token" {
		t.Fatal("personal token must remain unchanged")
	}
}
