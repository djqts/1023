// Package engine 加载外置 YAML 规则并在内存中执行匹配。
package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"banner-fingerprint/internal/domain/fingerprint"
)

type fileSpec struct {
	Version string     `yaml:"version"`
	Rules   []ruleSpec `yaml:"rules"`
}

type ruleSpec struct {
	ID             string   `yaml:"id"`
	Protocol       string   `yaml:"protocol"`
	Product        string   `yaml:"product"`
	Priority       int      `yaml:"priority"`
	Patterns       []string `yaml:"patterns"`
	OSHintPatterns []string `yaml:"os_hint_patterns"`
	OSHint         string   `yaml:"os_hint"`
	Confidence     float64  `yaml:"confidence"`
	Ports          []int    `yaml:"ports"`
}

type compiledRule struct {
	spec       ruleSpec
	patterns   []*regexp.Regexp
	osPatterns []*regexp.Regexp
}

// Engine 是加载完成的不可变规则集。
type Engine struct {
	rules  []compiledRule
	digest string
	count  int
}

// Load 从目录读取 *.yaml / *.yml，校验并编译规则。
func Load(dir string) (*Engine, error) {
	files, err := ruleFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no rule files under %q", dir)
	}

	var specs []ruleSpec
	seen := make(map[string]struct{})
	hash := sha256.New()

	for _, name := range files {
		content, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read rule file %s: %w", name, err)
		}
		hash.Write([]byte(filepath.Base(name)))
		hash.Write([]byte{0})
		hash.Write(content)
		hash.Write([]byte{0})

		var spec fileSpec
		if err := yaml.Unmarshal(content, &spec); err != nil {
			return nil, fmt.Errorf("parse rule file %s: %w", name, err)
		}
		for _, rule := range spec.Rules {
			if rule.ID == "" {
				return nil, fmt.Errorf("%s: rule id is empty", name)
			}
			if _, ok := seen[rule.ID]; ok {
				return nil, fmt.Errorf("%s: duplicate rule id %q", name, rule.ID)
			}
			seen[rule.ID] = struct{}{}
			if rule.Protocol == "" {
				return nil, fmt.Errorf("%s: rule %q protocol is empty", name, rule.ID)
			}
			if len(rule.Patterns) == 0 {
				return nil, fmt.Errorf("%s: rule %q has no patterns", name, rule.ID)
			}
			if rule.Confidence < 0 || rule.Confidence > 1 {
				return nil, fmt.Errorf("%s: rule %q confidence out of range", name, rule.ID)
			}
			specs = append(specs, rule)
		}
	}

	compiled := make([]compiledRule, 0, len(specs))
	for _, spec := range specs {
		rule := compiledRule{spec: spec}
		for _, pattern := range spec.Patterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("rule %q pattern %q: %w", spec.ID, pattern, err)
			}
			rule.patterns = append(rule.patterns, re)
		}
		for _, pattern := range spec.OSHintPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("rule %q os pattern %q: %w", spec.ID, pattern, err)
			}
			rule.osPatterns = append(rule.osPatterns, re)
		}
		compiled = append(compiled, rule)
	}

	sort.SliceStable(compiled, func(i, j int) bool {
		if compiled[i].spec.Priority != compiled[j].spec.Priority {
			return compiled[i].spec.Priority > compiled[j].spec.Priority
		}
		return compiled[i].spec.ID < compiled[j].spec.ID
	})

	return &Engine{
		rules:  compiled,
		digest: hex.EncodeToString(hash.Sum(nil)),
		count:  len(compiled),
	}, nil
}

func ruleFiles(dir string) ([]string, error) {
	var files []string
	for _, pattern := range []string{"*.yaml", "*.yml"} {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, fmt.Errorf("glob rules %s: %w", pattern, err)
		}
		files = append(files, matches...)
	}
	sort.Strings(files)
	return files, nil
}

// Fingerprint 对一条输入执行规则匹配；无命中时返回 unknown，不返回错误。
func (e *Engine) Fingerprint(in fingerprint.Input) (fingerprint.Result, error) {
	if e == nil || len(in.Banner) == 0 {
		return unknown(in), nil
	}

	banner := string(in.Banner)
	for _, rule := range e.rules {
		for _, re := range rule.patterns {
			match := re.FindStringSubmatch(banner)
			if match == nil {
				continue
			}

			version := ""
			if idx := re.SubexpIndex("version"); idx >= 0 && idx < len(match) {
				version = strings.TrimSpace(match[idx])
			}

			osHint := rule.spec.OSHint
			if osHint == "" {
				for _, osRe := range rule.osPatterns {
					osMatch := osRe.FindStringSubmatch(banner)
					if osMatch == nil {
						continue
					}
					if len(osMatch) > 1 && osMatch[1] != "" {
						osHint = osMatch[1]
					} else if len(osMatch) > 0 {
						osHint = osMatch[0]
					}
					osHint = strings.TrimSpace(osHint)
					break
				}
			}

			confidence := rule.spec.Confidence
			if confidence == 0 {
				confidence = 0.5
			}
			if confidence < 0 {
				confidence = 0
			}
			if confidence > 1 {
				confidence = 1
			}

			return fingerprint.Result{
				IP:         in.IP,
				Port:       in.Port,
				Protocol:   fingerprint.Protocol(rule.spec.Protocol),
				Product:    rule.spec.Product,
				Version:    version,
				OSHint:     osHint,
				Confidence: confidence,
			}, nil
		}
	}

	return unknown(in), nil
}

func unknown(in fingerprint.Input) fingerprint.Result {
	return fingerprint.Result{
		IP:         in.IP,
		Port:       in.Port,
		Protocol:   fingerprint.ProtocolUnknown,
		Confidence: 0,
	}
}

// RuleCount 返回已加载的规则数。
func (e *Engine) RuleCount() int {
	if e == nil {
		return 0
	}
	return e.count
}

// Digest 返回规则文件内容摘要。
func (e *Engine) Digest() string {
	if e == nil {
		return ""
	}
	return e.digest
}
