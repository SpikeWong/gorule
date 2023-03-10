package gorule

import (
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/spikewong/gorule/internal/parser"
)

func TestNewEngine(t *testing.T) {
	type args struct {
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want *Engine
	}{
		{
			name: "initialize engine",
			args: args{opts: []Option{}},
			want: &Engine{
				mu:     sync.Mutex{},
				rules:  make(map[string]*Rule, 0),
				config: &Config{SkipBadRuleDuringMatch: false},
				logger: log.New(os.Stdout, "", log.LstdFlags),
			},
		},
		{
			name: "initialize engine with opts",
			args: args{opts: []Option{
				WithConfig(&Config{SkipBadRuleDuringMatch: true}),
				WithLogger(log.New(io.Discard, "", log.LstdFlags)),
			}},
			want: &Engine{
				mu:     sync.Mutex{},
				rules:  make(map[string]*Rule, 0),
				config: &Config{SkipBadRuleDuringMatch: true},
				logger: log.New(io.Discard, "", log.LstdFlags),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEngine(tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngine_AddRule(t *testing.T) {
	type fields struct {
		mu     sync.Mutex
		rules  map[string]*Rule
		config *Config
		logger *log.Logger
	}
	type args struct {
		rule *Rule
	}

	rule := NewRule("test_rule@v1", "x + y > 2", func(i interface{}) (interface{}, error) {
		return "matched", nil
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  make(map[string]*Rule),
				config: &Config{SkipBadRuleDuringMatch: true},
				logger: log.New(io.Discard, "", log.LstdFlags),
			},
			args: args{
				rule: rule,
			},
			wantErr: false,
		},
		{
			name: "error: rule name exists",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  map[string]*Rule{rule.Name(): rule},
				config: &Config{SkipBadRuleDuringMatch: true},
				logger: log.New(io.Discard, "", log.LstdFlags),
			},
			args:    args{rule: rule},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				mu:     tt.fields.mu,
				rules:  tt.fields.rules,
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			if err := e.AddRule(tt.args.rule); (err != nil) != tt.wantErr {
				t.Errorf("AddRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngine_MatchWithoutFunctions(t *testing.T) {
	type fields struct {
		mu     sync.Mutex
		rules  map[string]*Rule
		config *Config
		logger *log.Logger
	}
	type args struct {
		vars      map[string]interface{}
		functions map[string]parser.ExpressionFunction
	}

	badGradeRule := NewRule(
		"if grade less than 40, then inform the parents",
		"grade < 40",
		func(i interface{}) (interface{}, error) {
			return "bad", nil
		})
	passedGradeRule := NewRule(
		"If grade gt 60, then passed",
		"grade > 60",
		func(i interface{}) (interface{}, error) {
			return "passed", nil
		})
	gradeRules := map[string]*Rule{
		badGradeRule.Name():    badGradeRule,
		passedGradeRule.Name(): passedGradeRule,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Rule
		wantErr bool
	}{
		{
			name: "happy path, matched one rule",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  gradeRules,
				config: &Config{SkipBadRuleDuringMatch: false},
				logger: log.New(os.Stdout, "", log.LstdFlags),
			},
			args:    args{vars: map[string]interface{}{"grade": 30}},
			wantErr: false,
			want:    []Rule{*badGradeRule},
		},
		{
			name: "happy path, not matched any rule",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  gradeRules,
				config: &Config{SkipBadRuleDuringMatch: false},
				logger: log.New(os.Stdout, "", log.LstdFlags),
			},
			args:    args{vars: map[string]interface{}{"grade": 50}},
			wantErr: false,
			want:    []Rule{},
		},
		{
			name: "error: missing vars",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  gradeRules,
				config: &Config{SkipBadRuleDuringMatch: false},
				logger: log.New(os.Stdout, "", log.LstdFlags),
			},
			args:    args{vars: map[string]interface{}{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				mu:     tt.fields.mu,
				rules:  tt.fields.rules,
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			got, err := e.Match(tt.args.vars, tt.args.functions)
			if (err != nil) != tt.wantErr {
				t.Errorf("Match() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("the length of got and want are not euqal, got: %d, want: %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v.Name() != tt.want[i].Name() {
					t.Errorf("matched rule not equal, got: %s, want: %s", v.Name(), tt.want[i].Name())
					return
				}
			}
		})
	}
}

func TestEngine_MatchWithFunctions(t *testing.T) {
	type fields struct {
		mu     sync.Mutex
		rules  map[string]*Rule
		config *Config
		logger *log.Logger
	}
	type args struct {
		vars      map[string]interface{}
		functions map[string]parser.ExpressionFunction
	}

	regexRule := NewRule(
		"match string based on regex",
		`matches(text, regex)`,
		func(i interface{}) (interface{}, error) {
			return "matched", nil
		},
	)
	goodFunctions := map[string]parser.ExpressionFunction{
		"matches": func(args ...interface{}) (interface{}, error) {
			return regexp.MustCompile(args[1].(string)).MatchString(args[0].(string)), nil
		},
	}
	badFunctions := map[string]parser.ExpressionFunction{
		"matches": func(args ...interface{}) (interface{}, error) {
			return nil, errors.New("bad function")
		},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Rule
		wantErr bool
	}{
		{
			name: "good function",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  map[string]*Rule{regexRule.Name(): regexRule},
				config: &Config{SkipBadRuleDuringMatch: false},
				logger: log.New(os.Stdout, "", log.LstdFlags),
			},
			args: args{
				vars: map[string]interface{}{
					"text":  "hello world",
					"regex": "hello.*",
				},
				functions: goodFunctions,
			},
			want:    []Rule{*regexRule},
			wantErr: false,
		},
		{
			name: "bad function",
			fields: fields{
				mu:     sync.Mutex{},
				rules:  map[string]*Rule{regexRule.Name(): regexRule},
				config: &Config{SkipBadRuleDuringMatch: false},
				logger: log.New(os.Stdout, "", log.LstdFlags),
			},
			args: args{
				vars: map[string]interface{}{
					"text":  "hello world",
					"regex": "hello.*",
				},
				functions: badFunctions,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				mu:     tt.fields.mu,
				rules:  tt.fields.rules,
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			got, err := e.Match(tt.args.vars, tt.args.functions)
			if (err != nil) != tt.wantErr {
				t.Errorf("Match() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("the length of got and want are not euqal, got: %d, want: %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v.Name() != tt.want[i].Name() {
					t.Errorf("matched rule not equal, got: %s, want: %s", v.Name(), tt.want[i].Name())
					return
				}
			}
		})
	}
}
