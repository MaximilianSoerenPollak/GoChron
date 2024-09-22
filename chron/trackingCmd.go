package chron

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var trackingCmd = &cobra.Command{
	Use:   "tracking",
	Short: "Currently tracking activity",
	Long:  "Show currently tracking activity.",
	Run: func(cmd *cobra.Command, args []string) {

		entry, err := database.GetRunningEntry()
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				fmt.Printf("%s No task currently running.", CharFinish)
				os.Exit(1)
			default:
				fmt.Printf("something went wrong getting current runnign entries. Error: %s", err.Error())
				os.Exit(1)
			}

		}
		// Kind of is redundant but I keep it for now.
		if entry == nil {
			fmt.Printf("%s No task currently running.", CharFinish)
			os.Exit(1)
		}
		fmt.Printf("%s %s", CharTrack, entry.GetOutputStrShort())
	},
}

func init() {
	rootCmd.AddCommand(trackingCmd)

	var err error
	database, err = InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
