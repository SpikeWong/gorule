package parser

import (
	"runtime"
)

func Evaluate(str string, variables map[string]interface{}) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	lexer := NewLexer(str, variables)
	yyNewParser().Parse(lexer)
	return lexer.Result(), nil
}
