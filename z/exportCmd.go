package z

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"encoding/csv"

	"github.com/jinzhu/now"
	"github.com/spf13/cobra"
)

func exportZeitJson(entries []Entry) (string, error) {
	stringified, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}

	return string(stringified), nil
}

func exportTymeJson(entries []Entry) (string, error) {
	tyme := Tyme{}
	err := tyme.FromEntries(entries)
	if err != nil {
		return "", err
	}

	return tyme.Stringify(), nil
}


func exportCSV(entries []Entry) error { 
	// TODO: There has to be a nicer way to write this, for now this is okay.
	f, err := os.OpenFile(fileName, os.O_CREATE| os.O_APPEND | os.O_WRONLY, 0666)
	if err != nil {
		return err 
	}
	defer f.Close()
	csvWriter := csv.NewWriter(f)
	if exportAllFields { 
		err := csvWriter.Write(entries[0].GetCSVHeaderAllData())
		if err != nil {
			return err 
		}
	} else {
		err := csvWriter.Write(entries[0].GetCSVHeaderShortData())
		if err != nil {
			return err 
		}
	}
	for _, v := range entries {
		if exportAllFields {
			err := csvWriter.Write(v.ConvertToCSVAllData())
			if err != nil {
				return err 
			}
		} else {
			err := csvWriter.Write(v.ConvertToCSVShortData())
			if err != nil {
				return err 
			}
		}
	}
	return nil 
}

var exportCmd = &cobra.Command{
	Use:   "export ([flags])",
	Short: "Export tracked activities",
	Long:  "Export tracked activities to various formats.",
	// Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entries []Entry
		var err error

		user := GetCurrentUser()

		entries, err = database.ListEntries(user)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}

		var sinceTime time.Time
		var untilTime time.Time

		if since != "" {
			sinceTime, err = now.Parse(since)
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		}

		if until != "" {
			untilTime, err = now.Parse(until)
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		}

		var filteredEntries []Entry
		filteredEntries, err = GetFilteredEntries(entries, project, task, sinceTime, untilTime)
		if err != nil {
			fmt.Printf("%s %+v\n", CharError, err)
			os.Exit(1)
		}

		if exportHours || exportDate {
			var addedInformationEntries []Entry
			for _, v := range filteredEntries {
				if exportHours {
					v.Hours = fmtHours(v.GetDuration())
				}
				if exportDate {
					v.SetDateFromBegining()
				}
				addedInformationEntries = append(addedInformationEntries, v)
			}
			// Reasignment here so we don't need to check other flags later
			filteredEntries = addedInformationEntries
		}

		var output = ""
		switch format {
		case "csv": 
			err = exportCSV(filteredEntries)
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		case "zeit":
			output, err = exportZeitJson(filteredEntries)
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		case "tyme":
			output, err = exportTymeJson(filteredEntries)
			if err != nil {
				fmt.Printf("%s %+v\n", CharError, err)
				os.Exit(1)
			}
		default:
			fmt.Printf("%s specify an export format; see `zeit export --help` for more info\n", CharError)
			os.Exit(1)
		}

		fmt.Printf("%s\n", output)
	},
}



func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&format, "format", "zeit", "Format to export, possible values: zeit, tyme")
	exportCmd.Flags().StringVar(&since, "since", "", "Date/time to start the export from")
	exportCmd.Flags().StringVar(&until, "until", "", "Date/time to export until")
	exportCmd.Flags().StringVarP(&project, "project", "p", "", "Project to be exported")
	exportCmd.Flags().StringVarP(&task, "task", "t", "", "Task to be exported")
	exportCmd.Flags().BoolVar(&exportDate, "date", true, "Set to true, if you want to export the 'Date' aswell")
	exportCmd.Flags().BoolVar(&exportHours, "hours-decimal", true, "Set to true if you want calculated Hours to be exported too")
	exportCmd.Flags().StringVar(&fileName, "file-name", "", "Set the output file for the csv export")
	exportCmd.Flags().BoolVar(&exportAllFields, "export-all-fields", false, "Set to true if you want to export all the available fields to the csv")

	var err error
	database, err = InitDatabase()
	if err != nil {
		fmt.Printf("%s %+v\n", CharError, err)
		os.Exit(1)
	}
}
