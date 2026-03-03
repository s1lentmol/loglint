package analyzer_test

import (
	"path/filepath"
	"testing"

	"github.com/s1lentmol/loglint/internal/analyzer"
	"github.com/s1lentmol/loglint/internal/config"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := mustTestdataPath(t)

	analysistest.Run(
		t,
		testdata,
		analyzer.Analyzer,
		"valid",
		"lowercase",
		"english",
		"specialchars",
		"sensitive",
		"mixed",
		"edgecases",
	)
}

func TestAnalyzer_ConfigDisablesSpecialCharsRule(t *testing.T) {
	t.Parallel()

	testdata := mustTestdataPath(t)

	cfg := config.Default()
	cfg.Rules.NoSpecialChars = false

	an, err := analyzer.New(cfg)
	if err != nil {
		t.Fatalf("create analyzer: %v", err)
	}

	analysistest.Run(t, testdata, an, "config_disabled_special")
}

func TestAnalyzer_ConfigSensitiveOverride(t *testing.T) {
	t.Parallel()

	testdata := mustTestdataPath(t)

	cfg := config.Default()
	cfg.Sensitive.Mode = config.SensitiveModeOverride
	cfg.Sensitive.Keywords = []string{"sessionid"}

	an, err := analyzer.New(cfg)
	if err != nil {
		t.Fatalf("create analyzer: %v", err)
	}

	analysistest.Run(t, testdata, an, "config_sensitive_override")
}

func TestAnalyzer_ConfigIgnorePath(t *testing.T) {
	t.Parallel()

	testdata := mustTestdataPath(t)

	cfg := config.Default()
	cfg.Ignore.Paths = []string{"**/config_ignore_path/*.go"}

	an, err := analyzer.New(cfg)
	if err != nil {
		t.Fatalf("create analyzer: %v", err)
	}

	analysistest.Run(t, testdata, an, "config_ignore_path")
}

func mustTestdataPath(t *testing.T) string {
	t.Helper()

	testdata, err := filepath.Abs(filepath.Join("..", "..", "testdata"))
	if err != nil {
		t.Fatalf("resolve testdata path: %v", err)
	}

	return testdata
}
