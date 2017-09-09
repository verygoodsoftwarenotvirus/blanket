package main

import (
	"fmt"
	"os"

	"github.com/verygoodsoftwarenotvirus/tarp/cmd"
)

var (
	verbose bool = true
)

func main() {
	if err := tarp.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
