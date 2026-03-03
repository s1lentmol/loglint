package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := Default()
	if cfg.Version != 1 {
		t.Fatalf("Version = %d, want 1", cfg.Version)
	}
	if !cfg.Rules.LowercaseStart || !cfg.Rules.EnglishOnly || !cfg.Rules.NoSpecialChars || !cfg.Rules.NoSensitiveData {
		t.Fatalf("all rules must be enabled by default")
	}
	if cfg.Sensitive.Mode != SensitiveModeAppend {
		t.Fatalf("Sensitive.Mode = %q, want %q", cfg.Sensitive.Mode, SensitiveModeAppend)
	}
}

func TestLoad_AutodiscoveryMissing(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Version != 1 {
		t.Fatalf("Version = %d, want 1", cfg.Version)
	}
}

func TestLoad_InvalidCases(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	write := func(name, content string) string {
		p := filepath.Join(tmp, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		return p
	}

	invalidYAML := write("invalid.yml", "version: [")
	if _, err := Load(invalidYAML); err == nil {
		t.Fatalf("expected yaml error")
	}

	invalidVersion := write("invalid-version.yml", "version: 2\n")
	if _, err := Load(invalidVersion); err == nil {
		t.Fatalf("expected version error")
	}

	invalidMode := write("invalid-mode.yml", "version: 1\nsensitive:\n  mode: bad\n")
	if _, err := Load(invalidMode); err == nil {
		t.Fatalf("expected sensitive.mode error")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	path := filepath.Join(tmp, ".loglint.yml")
	content := `version: 1
rules:
  lowercase_start: false
  english_only: true
  no_special_chars: false
  no_sensitive_data: true
sensitive:
  mode: override
  keywords:
    - SESSIONID
    - private_key
    - private_key
ignore:
  paths:
    - "vendor/**"
    - "**/*_generated.go"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Rules.LowercaseStart {
		t.Fatalf("LowercaseStart = true, want false")
	}
	if cfg.Rules.NoSpecialChars {
		t.Fatalf("NoSpecialChars = true, want false")
	}
	if cfg.Sensitive.Mode != SensitiveModeOverride {
		t.Fatalf("Sensitive.Mode = %q, want %q", cfg.Sensitive.Mode, SensitiveModeOverride)
	}
	if len(cfg.Sensitive.Keywords) != 2 {
		t.Fatalf("len(Sensitive.Keywords) = %d, want 2", len(cfg.Sensitive.Keywords))
	}
	if len(cfg.Ignore.Paths) != 2 {
		t.Fatalf("len(Ignore.Paths) = %d, want 2", len(cfg.Ignore.Paths))
	}
}

func TestSensitiveAppendOverride(t *testing.T) {
	t.Parallel()

	def := Default()
	appended := MergeWithDefault(Config{
		Rules: def.Rules,
		Sensitive: SensitiveConfig{
			Mode:     SensitiveModeAppend,
			Keywords: []string{"sessionid"},
		},
	})
	if appended.Sensitive.Mode != SensitiveModeAppend {
		t.Fatalf("append mode lost")
	}

	overridden := MergeWithDefault(Config{
		Rules: def.Rules,
		Sensitive: SensitiveConfig{
			Mode:     SensitiveModeOverride,
			Keywords: []string{"sessionid"},
		},
	})
	if overridden.Sensitive.Mode != SensitiveModeOverride {
		t.Fatalf("override mode lost")
	}
}

func TestIsIgnored(t *testing.T) {
	t.Parallel()

	patterns := []string{"vendor/**", "**/*_generated.go"}
	if !IsIgnored("vendor/a/b/c.go", patterns) {
		t.Fatalf("expected vendor path to be ignored")
	}
	if !IsIgnored("internal/foo_generated.go", patterns) {
		t.Fatalf("expected generated file to be ignored")
	}
	if IsIgnored("internal/foo.go", patterns) {
		t.Fatalf("did not expect ordinary file to be ignored")
	}
}
