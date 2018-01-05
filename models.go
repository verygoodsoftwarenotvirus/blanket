package main

import (
	"go/token"

	"github.com/fatih/set"
)

type blanketOutput struct {
	DeclaredCount             int                      `json:"declared"`
	CalledCount               int                      `json:"called"`
	Score                     int                      `json:"score"`
	Details                   map[string][]blanketFunc `json:"-"`
	LongestFunctionNameLength int                      `json:"-"`
}

type blanketReport struct {
	DeclaredDetails map[string]blanketFunc
	Called          *set.Set
	Declared        *set.Set
}

type blanketDetails []blanketFunc
type blanketFunc struct {
	Name      string
	Filename  string
	DeclPos   token.Position
	RBracePos token.Position
	LBracePos token.Position
}

func (td blanketDetails) Len() int {
	return len(td)
}

func (td blanketDetails) Less(i, j int) bool {
	if td[i].Filename < td[j].Filename {
		return true
	}
	if td[i].Filename > td[j].Filename {
		return false
	}
	return td[i].DeclPos.Line < td[j].DeclPos.Line
}

func (td blanketDetails) Swap(i, j int) {
	td[i], td[j] = td[j], td[i]
}
