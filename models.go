package main

import (
	"go/token"

	"github.com/fatih/set"
)

type TarpOutput struct {
	DeclaredCount int
	CalledCount   int
	Score         int
	Details       map[string][]TarpFunc
	LongestFunctionNameLength int
}

type TarpReport struct {
	DeclaredDetails map[string]TarpFunc
	Called          *set.Set
	Declared        *set.Set
}

type TarpDetails []TarpFunc
type TarpFunc struct {
	Name      string
	Filename  string
	DeclPos   token.Position
	RBracePos token.Position
	LBracePos token.Position
}

func (td TarpDetails) Len() int {
	return len(td)
}

func (td TarpDetails) Less(i, j int) bool {
	if td[i].Filename < td[j].Filename {
		return true
	}
	if td[i].Filename > td[j].Filename {
		return false
	}
	return td[i].DeclPos.Line < td[j].DeclPos.Line
}

func (td TarpDetails) Swap(i, j int) {
	td[i], td[j] = td[j], td[i]
}
