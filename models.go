package main

import (
	"go/token"

	"github.com/fatih/set"
)

type tarpOutput struct {
	DeclaredCount             int
	CalledCount               int
	Score                     int
	Details                   map[string][]tarpFunc
	LongestFunctionNameLength int
}

type tarpReport struct {
	DeclaredDetails map[string]tarpFunc
	Called          *set.Set
	Declared        *set.Set
}

type tarpDetails []tarpFunc
type tarpFunc struct {
	Name      string
	Filename  string
	DeclPos   token.Position
	RBracePos token.Position
	LBracePos token.Position
}

func (td tarpDetails) Len() int {
	return len(td)
}

func (td tarpDetails) Less(i, j int) bool {
	if td[i].Filename < td[j].Filename {
		return true
	}
	if td[i].Filename > td[j].Filename {
		return false
	}
	return td[i].DeclPos.Line < td[j].DeclPos.Line
}

func (td tarpDetails) Swap(i, j int) {
	td[i], td[j] = td[j], td[i]
}
