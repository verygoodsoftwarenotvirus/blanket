package util

import (
	"fmt"
	"os"
	"testing"
)

func TestBuildExamplePackagePath(t *testing.T) {
	t.Parallel()

	result := BuildExamplePackagePath(t, "example_packages/simple", false)
	if result != "github.com/verygoodsoftwarenotvirus/blanket/example_packages/example_packages/simple" {
		t.Logf("Expected '%s', got '%s'", "", result)
		t.Fail()
	}
}

func TestBuildExampleFilePath(t *testing.T) {
	t.Parallel()

	result := BuildExampleFilePath("example_packages/simple")
	if result != fmt.Sprintf("%s/src/github.com/verygoodsoftwarenotvirus/blanket/example_files/example_packages/simple", os.Getenv("GOPATH")) {
		t.Logf("Expected '%s', got '%s'", "", result)
		t.Fail()
	}
}
