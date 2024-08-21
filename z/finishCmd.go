package z

import (
	"database/sql"
	"errors"
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
			switch {
			case errors.Is(err, sql.ErrNoRows):
				fmt.Printf("%s no task is currently running. Can only finish a running task.\n", CharFinish)
				os.Exit(1)
			default:
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		}
		// Redudant but I keep it for now.
		if runningEntry == nil {
			fmt.Printf("%s no task is currently running. Can only finish a running task.\n", CharFinish)
			os.Exit(1)
		}
		// Finishing the entry
		runningEntry.SetFinish()
		if notes != "" {
			runningEntry.Notes = strings.ReplaceAll(notes, "\\n", "\n")
		}
		err = database.AddFinishToEntry(*runningEntry)
		if err != nil {
			fmt.Printf("%s something ent wrong updating the entry. Error: %s", CharError, err.Error())
			os.Exit(1)
		}
		fmt.Println(runningEntry.GetOutputForFinish())
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
