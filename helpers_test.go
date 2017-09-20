package main

import (
	"os"
	"strings"
	"testing"
)

////////////////////////////////////////////////////////
//                                                    //
//               Test Helper Functions                //
//                                                    //
////////////////////////////////////////////////////////

type subtest struct {
	Message string
	Test    func(t *testing.T)
}

func runSubtestSuite(t *testing.T, tests []subtest) {
	t.Helper()
	testPassed := true
	for _, test := range tests {
		if !testPassed {
			t.FailNow()
		}
		testPassed = t.Run(test.Message, test.Test)
	}
}

func buildExamplePackagePath(t *testing.T, packageName string, abs bool) string {
	t.Helper()
	gopath := os.Getenv("GOPATH")
	if abs {
		return strings.Join([]string{gopath, "src", "github.com", "verygoodsoftwarenotvirus", "tarp", "example_packages", packageName}, "/")
	}
	return strings.Join([]string{"github.com", "verygoodsoftwarenotvirus", "tarp", "example_packages", packageName}, "/")
}
