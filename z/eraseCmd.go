package z

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

var eraseCmd = &cobra.Command{
	Use:   "erase ([flags]) [id]",
	Short: "Erase activity",
	Long:  "Erase tracked activity.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		idStr := args[0]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Printf("%s %s", CharError, "Please provide a valid number")
			os.Exit(1)
		}

		err = database.DeleteEntry(int64(id))
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}

		fmt.Printf("%s erased %s\n", CharInfo, color.FgLightWhite.Render(id))
	},
}

func init() {
	rootCmd.AddCommand(eraseCmd)

	var err error
	database, err = InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
