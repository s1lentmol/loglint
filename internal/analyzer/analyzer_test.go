package analyzer_test

import (
	"path/filepath"
	"testing"

	"github.com/s1lentmol/loglint/internal/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata, err := filepath.Abs(filepath.Join("..", "..", "testdata"))
	if err != nil {
		t.Fatalf("resolve testdata path: %v", err)
	}

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
