package main

import (
	"go/token"
)

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
