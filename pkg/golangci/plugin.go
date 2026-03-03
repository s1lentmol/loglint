package golangci

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/s1lentmol/loglint/internal/analyzer"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("loglint", New)
}

func New(conf any) (register.LinterPlugin, error) {
	_ = conf
	return &Plugin{}, nil
}

type Plugin struct{}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.Analyzer}, nil
}

func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
