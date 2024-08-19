package z

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import ([flags]) [file]",
	Short: "Import tracked activities",
	Long:  "Import tracked activities from various formats.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// TODO:
		fmt.Printf("%s not yet implemented\n", CharError)
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&format, "format", "zeit", "Format to import, possible values: zeit, tyme")

	var err error
	database, err = InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
