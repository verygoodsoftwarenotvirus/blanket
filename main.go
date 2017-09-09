package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// "github.com/verygoodsoftwarenotvirus/tarp/cmd"
)

var (
	failOnFinding  bool
	analyzePackage string
)

var RootCmd = &cobra.Command{
	Use:   "tarp",
	Short: "tarp is a test coverage helper",
	Long:  `tarp is a test coverage helper`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// log.Println("RootCmd.Run called")
	// },
}

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a given package",
	Long:  "Analyze takes a given package's code and determines which functions lack direct test coverage.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if analyzePackage == "" {
			return errors.New("the required flag `-p, --package` was not specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		analyze(analyzePackage, failOnFinding)
	},
}

func init() {
	RootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&failOnFinding, "fail-on-finding", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", "", "Package to run analyze on")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
