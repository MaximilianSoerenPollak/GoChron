package z

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [flags] 'importFile'",
	Short: "Import tracked activities",
	Long: `Import tracked activities from various formats.

zeit: output from -> 'zeit export --format "zeit"'. 
csv: output from -> 'zeit export --format "csv" --export-all-fields'.

For the 'csv' output, please make sure you export all the fields.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		importFile := args[0]
		if format == "" {
			fmt.Printf("%s Please specify a format string. 'zeit', 'csv-short', 'csv-long'\n", CharError)
			os.Exit(1)
		}
		if importFile == "" {
			fmt.Printf("%s Please specify an import file.\n", CharError)
			os.Exit(1)
		}
		switch format {
		case "zeit":
			fileContent, err := os.ReadFile(importFile)
			if err != nil {
				fmt.Printf("%s The import file could not be read. Error: %s\n", CharError, err.Error())
				os.Exit(1)
			}
			var entries []EntryDB
			err = json.Unmarshal(fileContent, &entries)
			if err != nil {
				fmt.Printf("%s Could not unmarshal file into 'entryDB' struct. Error: %s\n", CharError, err.Error())
				os.Exit(1)
			}
			for _, v := range entries {
				// Hack to make it work, probably should fix this someday
				v.ID = "1"
				v.Running = false
				entryConv, err := v.ConvertToEntry()
				if err != nil {
					fmt.Printf("%s Could not convert entryDB '%+v' to entry. Error: %s\n", CharError, v, err.Error())
					os.Exit(1)
				}
				err = database.AddEntry(entryConv, false)
				if err != nil {
					fmt.Printf("%s Could not add entry '%+v' to the database. Error: %s\n", CharError, entryConv, err.Error())
					os.Exit(1)
				}
				if verbose {
					fmt.Printf("%s added Entry: '%s' to the database\n", CharInfo, entryConv.GetOutputStrShort())
				}
			}
			fmt.Printf("%s added all entries to the database\n", CharInfo)

		case "csv":
			file, err := os.Open(importFile)
			if err != nil {
				fmt.Printf("%s The import file could not be opened. Error: %s\n", CharError, err.Error())
				os.Exit(1)
			}
			csvReader := csv.NewReader(file)
			csvReader.Comma = ';'
			content, err := csvReader.ReadAll()
			if err != nil {
				fmt.Printf("%s Something went wrong reading the import file. Error: %s\n", CharError, err.Error())
				os.Exit(1)
			}
			// We are skipping the header
			for _, v := range content[1:] {
				if len(v) != 7{
					fmt.Printf("%s The csv has the wrong format, please make sure you export the csv with all fields\n", CharError)
					fmt.Printf("%s The fields expected in THIS order: date, start, finish, project, task, hours, notes\n", CharInfo)
					os.Exit(1)
				}
				var eDB EntryDB
				// Hack to make it work, probably should fix this someday
				eDB.ID = "1"
				eDB.Date = v[0]
				eDB.Begin = v[1]
				eDB.Finish = v[2]
				eDB.Project = v[3]
				eDB.Task = v[4]
				eDB.Hours = v[5]
				eDB.Notes = v[6]
				eDB.Running = false
				entryConv, err := eDB.ConvertToEntry()
				if err != nil {
					fmt.Printf("%s Could not convert entryDB '%+v' to entry. Error: %s\n", CharError, eDB, err.Error())
					os.Exit(1)
				}
				err = database.AddEntry(entryConv, false)
				if err != nil {
					fmt.Printf("%s Could not add entry '%+v' to the database. Error: %s\n", CharError, entryConv, err.Error())
					os.Exit(1)
				}
				if verbose {
					fmt.Printf("%s added Entry: '%s' to the database\n", CharInfo, entryConv.GetOutputStrShort())
				}

			}
			fmt.Printf("%s added all entries to the database\n", CharInfo)

		default:
			fmt.Printf("%s Could not find an approved format. Please try again\n", CharError)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&format, "format", "", "Format to import, possible values: zeit, csv")
	importCmd.Flags().BoolVar(&verbose, "verbose", false, "Show output for each added entry. Default: false")

	var err error
	database, err = InitDB()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
