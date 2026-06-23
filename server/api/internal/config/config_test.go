package config

import "testing"

func TestValidateRuntimeDependenciesAllowsLocalDefaults(t *testing.T) {
	cfg := Config{AppEnv: "local"}
	if err := cfg.ValidateRuntimeDependencies(); err != nil {
		t.Fatalf("local config should not require deploy dependencies: %v", err)
	}
}

func TestValidateRuntimeDependenciesRequiresDeployDependencies(t *testing.T) {
	cfg := Config{AppEnv: "staging"}
	if err := cfg.ValidateRuntimeDependencies(); err == nil {
		t.Fatal("expected staging config without MySQL and storage to fail")
	}

	cfg.MySQLDSN = "memotree:secret@tcp(mysql:3306)/memotree?parseTime=true"
	if err := cfg.ValidateRuntimeDependencies(); err == nil {
		t.Fatal("expected staging config without storage credentials to fail")
	}

	cfg.StorageAccessKeyID = "key"
	cfg.StorageSecretKey = "secret"
	cfg.StorageEndpoint = "https://example.r2.cloudflarestorage.com"
	cfg.OriginalsBucket = "originals"
	cfg.PreviewsBucket = "previews"
	if err := cfg.ValidateRuntimeDependencies(); err != nil {
		t.Fatalf("expected complete staging config to pass: %v", err)
	}
}
