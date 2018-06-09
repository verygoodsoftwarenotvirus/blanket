package util

import (
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
