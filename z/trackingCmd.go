package z

import (
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
			fmt.Printf("something went wrong getting current runnign entries. Error: %s", err.Error())
			os.Exit(1)
		}
		if entry == nil {
			fmt.Printf("%s No task currently running.", CharFinish)
			os.Exit(1)
		}
		fmt.Printf("%s %s", CharTrack, entry.GetOutputStrShort())
		return
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
