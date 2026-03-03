package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultConfigFile = ".loglint.yml"

	SensitiveModeAppend   = "append"
	SensitiveModeOverride = "override"
)

type Config struct {
	Version   int             `yaml:"version" mapstructure:"version"`
	Rules     RulesConfig     `yaml:"rules" mapstructure:"rules"`
	Sensitive SensitiveConfig `yaml:"sensitive" mapstructure:"sensitive"`
	Ignore    IgnoreConfig    `yaml:"ignore" mapstructure:"ignore"`
}

type RulesConfig struct {
	LowercaseStart  bool `yaml:"lowercase_start" mapstructure:"lowercase_start"`
	EnglishOnly     bool `yaml:"english_only" mapstructure:"english_only"`
	NoSpecialChars  bool `yaml:"no_special_chars" mapstructure:"no_special_chars"`
	NoSensitiveData bool `yaml:"no_sensitive_data" mapstructure:"no_sensitive_data"`
}

type SensitiveConfig struct {
	Mode     string   `yaml:"mode" mapstructure:"mode"`
	Keywords []string `yaml:"keywords" mapstructure:"keywords"`
}

type IgnoreConfig struct {
	Paths []string `yaml:"paths" mapstructure:"paths"`
}

type rawConfig struct {
	Version   *int         `yaml:"version" mapstructure:"version"`
	Rules     rawRules     `yaml:"rules" mapstructure:"rules"`
	Sensitive rawSensitive `yaml:"sensitive" mapstructure:"sensitive"`
	Ignore    IgnoreConfig `yaml:"ignore" mapstructure:"ignore"`
}

type rawRules struct {
	LowercaseStart  *bool `yaml:"lowercase_start" mapstructure:"lowercase_start"`
	EnglishOnly     *bool `yaml:"english_only" mapstructure:"english_only"`
	NoSpecialChars  *bool `yaml:"no_special_chars" mapstructure:"no_special_chars"`
	NoSensitiveData *bool `yaml:"no_sensitive_data" mapstructure:"no_sensitive_data"`
}

type rawSensitive struct {
	Mode     *string  `yaml:"mode" mapstructure:"mode"`
	Keywords []string `yaml:"keywords" mapstructure:"keywords"`
}

func Default() Config {
	return Config{
		Version: 1,
		Rules: RulesConfig{
			LowercaseStart:  true,
			EnglishOnly:     true,
			NoSpecialChars:  true,
			NoSensitiveData: true,
		},
		Sensitive: SensitiveConfig{
			Mode:     SensitiveModeAppend,
			Keywords: nil,
		},
		Ignore: IgnoreConfig{
			Paths: nil,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	resolvedPath, found, err := resolvePath(path)
	if err != nil {
		return Config{}, err
	}
	if !found {
		return cfg, nil
	}

	v := viper.New()
	v.SetConfigFile(resolvedPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("loglint config: invalid yaml: %w", err)
	}

	var raw rawConfig
	if err := v.Unmarshal(&raw); err != nil {
		return Config{}, fmt.Errorf("loglint config: invalid yaml: %w", err)
	}

	return mergeRawWithDefault(cfg, raw)
}

func MergeWithDefault(cfg Config) Config {
	def := Default()

	if cfg.Version != 0 {
		def.Version = cfg.Version
	}
	def.Rules = cfg.Rules

	if cfg.Sensitive.Mode != "" {
		def.Sensitive.Mode = strings.ToLower(strings.TrimSpace(cfg.Sensitive.Mode))
	}
	def.Sensitive.Keywords = normalizeKeywords(cfg.Sensitive.Keywords)

	def.Ignore.Paths = normalizePatterns(cfg.Ignore.Paths)
	return def
}

func resolvePath(path string) (string, bool, error) {
	if strings.TrimSpace(path) != "" {
		resolved := filepath.Clean(path)
		if _, err := os.Stat(resolved); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", false, fmt.Errorf("loglint config: file not found: %s", resolved)
			}
			return "", false, fmt.Errorf("loglint config: stat %s: %w", resolved, err)
		}
		return resolved, true, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("loglint config: get cwd: %w", err)
	}

	auto := filepath.Join(cwd, DefaultConfigFile)
	if _, err := os.Stat(auto); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("loglint config: stat %s: %w", auto, err)
	}

	return auto, true, nil
}

func mergeRawWithDefault(def Config, raw rawConfig) (Config, error) {
	out := def

	if raw.Version == nil {
		return Config{}, fmt.Errorf("loglint config: version is required")
	}
	if *raw.Version != 1 {
		return Config{}, fmt.Errorf("loglint config: unknown version %d", *raw.Version)
	}
	out.Version = *raw.Version

	if raw.Rules.LowercaseStart != nil {
		out.Rules.LowercaseStart = *raw.Rules.LowercaseStart
	}
	if raw.Rules.EnglishOnly != nil {
		out.Rules.EnglishOnly = *raw.Rules.EnglishOnly
	}
	if raw.Rules.NoSpecialChars != nil {
		out.Rules.NoSpecialChars = *raw.Rules.NoSpecialChars
	}
	if raw.Rules.NoSensitiveData != nil {
		out.Rules.NoSensitiveData = *raw.Rules.NoSensitiveData
	}

	if raw.Sensitive.Mode != nil {
		out.Sensitive.Mode = strings.ToLower(strings.TrimSpace(*raw.Sensitive.Mode))
	}
	if out.Sensitive.Mode != SensitiveModeAppend && out.Sensitive.Mode != SensitiveModeOverride {
		return Config{}, fmt.Errorf("loglint config: invalid sensitive.mode %q", out.Sensitive.Mode)
	}
	if raw.Sensitive.Keywords != nil {
		out.Sensitive.Keywords = normalizeKeywords(raw.Sensitive.Keywords)
	}

	if raw.Ignore.Paths != nil {
		out.Ignore.Paths = normalizePatterns(raw.Ignore.Paths)
	}

	return out, nil
}

func normalizeKeywords(in []string) []string {
	dedup := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, kw := range in {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw == "" {
			continue
		}
		if _, ok := dedup[kw]; ok {
			continue
		}
		dedup[kw] = struct{}{}
		out = append(out, kw)
	}

	return out
}

func normalizePatterns(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))

	for _, p := range in {
		p = strings.TrimSpace(filepath.ToSlash(p))
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	return out
}

type IgnoreMatcher struct {
	regexps []*regexp.Regexp
}

func CompileIgnoreMatcher(patterns []string) (IgnoreMatcher, error) {
	regexps := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := globToRegexp(p)
		if err != nil {
			return IgnoreMatcher{}, fmt.Errorf("loglint config: invalid ignore pattern %q: %w", p, err)
		}
		regexps = append(regexps, re)
	}
	return IgnoreMatcher{regexps: regexps}, nil
}

func IsIgnored(path string, patterns []string) bool {
	m, err := CompileIgnoreMatcher(patterns)
	if err != nil {
		return false
	}
	return m.Match(path)
}

func (m IgnoreMatcher) Match(path string) bool {
	if len(m.regexps) == 0 {
		return false
	}

	p := filepath.ToSlash(path)
	for _, re := range m.regexps {
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

func globToRegexp(glob string) (*regexp.Regexp, error) {
	glob = strings.TrimSpace(filepath.ToSlash(glob))
	if glob == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	var b strings.Builder
	b.WriteString("^")

	for i := 0; i < len(glob); i++ {
		ch := glob[i]
		switch ch {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(ch)))
		}
	}

	pattern := b.String() + "$"
	if strings.HasPrefix(glob, "**/") {
		pattern = "^(?:.*/)?" + strings.TrimPrefix(pattern, "^.*/")
	}

	return regexp.Compile(pattern)
}
