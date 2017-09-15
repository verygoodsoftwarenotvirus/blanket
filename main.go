package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	failOnFinding  bool
	analyzePackage string
)

var rootCmd = &cobra.Command{
	Use:   "tarp",
	Short: "tarp is a temp coverage helper",
	Long:  `tarp is a temp coverage helper`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// log.Println("RootCmd.Run called")
	// },
}

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a given package",
	Long:  "Analyze takes a given package's code and determines which functions lack direct temp coverage.",
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
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().BoolVarP(&failOnFinding, "fail-on-found", "f", false, "Call os.Exit(1) when functions without direct tests are found")
	analyzeCmd.Flags().StringVarP(&analyzePackage, "package", "p", "", "Package to run analyze on")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
