package gorule

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/davecgh/go-spew/spew"

	"github.com/spikewong/gorule/internal/parser"
)

var (
	ErrRuleExists       = errors.New("rule name already exists")
	ErrNonBooleanResult = errors.New("encountered non boolean result during eval rule")
)

type Config struct {
	SkipBadRuleDuringMatch bool
}

type Engine struct {
	mu sync.Mutex

	rules  map[string]*Rule
	config *Config
	logger *log.Logger
}

type Option func(*Engine)

// NewEngine initializes engine with options
// by default SkipBadRuleDuringMatch is false, which means will return error if error
// or non-boolean value encountered during Match.
func NewEngine(opts ...Option) *Engine {
	engine := &Engine{
		rules:  make(map[string]*Rule),
		config: &Config{SkipBadRuleDuringMatch: false},
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	for _, opt := range opts {
		opt(engine)
	}

	return engine
}

// WithConfig sets config for engine.
func WithConfig(config *Config) Option {
	return func(e *Engine) {
		e.config = config
	}
}

// WithLogger sets logger for engine.
func WithLogger(logger *log.Logger) Option {
	return func(e *Engine) {
		e.logger = logger
	}
}

// AddRule adds rule into engine, return error if rule name exists.
func (e *Engine) AddRule(rule *Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for name, _ := range e.rules {
		if name == rule.Name() {
			return fmt.Errorf("%w: %s", ErrRuleExists, name)
		}
	}

	e.rules[rule.Name()] = rule

	return nil
}

// Match iterates through all the rules of the engine and will return the matching rules.
func (e *Engine) Match(vars map[string]interface{}, functions map[string]parser.ExpressionFunction) ([]Rule, error) {
	matchedRules := make([]Rule, 0)

	for _, r := range e.rules {
		res, err := parser.Evaluate(r.condition, vars, functions)
		matched, ok := res.(bool)
		if !e.config.SkipBadRuleDuringMatch {
			if err != nil {
				e.logger.Printf("Error: rule %s returned unexpected error during match: %v", r.Name(), err)
				return nil, fmt.Errorf("unexpected error occured during match: %w", err)
			} else if !ok {
				e.logger.Printf("Error: rule %s returned non-boolean value with vars %v", r.Name(), spew.Sdump(vars))
				return nil, fmt.Errorf("%s: %w", r.Name(), ErrNonBooleanResult)
			}
		}

		if matched {
			matchedRules = append(matchedRules, *r)
		}
	}

	return matchedRules, nil
}
