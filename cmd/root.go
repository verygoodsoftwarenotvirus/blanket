package tarp

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tarp",
	Short: "tarp is a test coverage helper",
	Long:  `tarp is a test coverage helper`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// log.Println("RootCmd.Run called")
	// },
}
