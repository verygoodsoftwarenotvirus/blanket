package util

import (
	"fmt"
	"os"
	"testing"
)

////////////////////////////////////////////////////////
//                                                    //
//               Test Helper Functions                //
//                                                    //
////////////////////////////////////////////////////////

func BuildExamplePackagePath(t *testing.T, packageName string, abs bool) string {
	t.Helper()
	gopath := os.Getenv("GOPATH")
	if abs {
		return fmt.Sprintf("%s/src/github.com/verygoodsoftwarenotvirus/blanket/example_packages/%s", gopath, packageName)
	}
	return fmt.Sprintf("github.com/verygoodsoftwarenotvirus/blanket/example_packages/%s", packageName)
}
