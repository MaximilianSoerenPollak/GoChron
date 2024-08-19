package z

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Finish currently running activity",
	Long:  "Finishing tracking of currently running activity.",
	Run: func(cmd *cobra.Command, args []string) {

		runningEntry, err := database.GetRunningEntry()
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}

		if runningEntry == nil {
			fmt.Printf("%s no task is currently running. Can only finish a running task.\n", CharFinish)
			os.Exit(1)
		}
		// Finishing the entry
		runningEntry.SetFinish()
		if notes != "" {
			runningEntry.Notes = strings.Replace(notes, "\\n", "\n", -1)
		}
		err = database.UpdateEntry(*runningEntry)
		if err != nil {
			fmt.Errorf("%s something ent wrong updating the entry. Error: %s", CharError, err.Error())
			os.Exit(1)
		}
		fmt.Println(runningEntry.GetOutputForFinish())
		return
	},
}

func init() {
	rootCmd.AddCommand(finishCmd)
	finishCmd.Flags().StringVarP(&finish, "finish", "s", "", "Time the activity should finish at\n\nEither in the formats 16:00 / 4:00PM \nor relative to the current time, \ne.g. -0:15 (now minus 15 minutes), +1.50 (now plus 1:30h).\nMust be after --begin time.")
	finishCmd.Flags().StringVarP(&notes, "notes", "n", "", "Add notes to the task while finishing it.")

	var err error
	database, err = InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
