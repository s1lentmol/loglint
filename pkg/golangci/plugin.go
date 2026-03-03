package golangci

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"github.com/s1lentmol/loglint/internal/analyzer"
	"github.com/s1lentmol/loglint/internal/config"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("loglint", New)
}

func New(conf any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](conf)
	if err != nil {
		return nil, fmt.Errorf("loglint config: decode plugin settings: %w", err)
	}

	cfg, err := config.Load(settings.Config)
	if err != nil {
		return nil, err
	}

	an, err := analyzer.New(cfg)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		analyzers: []*analysis.Analyzer{an},
	}, nil
}

type Settings struct {
	Config string `json:"config" yaml:"config" mapstructure:"config"`
}

type Plugin struct {
	analyzers []*analysis.Analyzer
}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return p.analyzers, nil
}

func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
