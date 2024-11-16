package chron

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
)

type Database struct {
	DB *sql.DB
}

// Decision: We do not care about the UUID and would rather use incremental ID to make also selecting easier.
// Also I would like to remove the 'user' from the equation as this is suppose to be a 'one user' CLI
func InitDB() (*Database, error) {
	// Will make '.config/zeit.db' the default
	dbLocation, ok := os.LookupEnv("ZEIT_DB")
	if !ok || dbLocation == "" {
		fmt.Println("Did not find 'ZEIT_DB' env. variable specified. Will use `$HOME/.config/zeit.db` as default")
		dbLocation = "$HOME/.config/zeit.db"
	}
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		fmt.Printf("Encountered error opening the db, Error: %s", err.Error())
		return nil, err
	}
	err = createDefaultTables(db)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

func (db *Database) AddEntry(entry *Entry, running bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{
		entry.Date,
		entry.Begin.Truncate(0).String(),
		entry.Finish.Truncate(0).String(),
		entry.Hours.String(),
		entry.Project,
		strings.ReplaceAll(entry.Task, `'`, `"`),
		entry.Notes,
		running}
	query := fmt.Sprintf(`INSERT INTO entries(date, start, finish, hours, project, task, notes, running) 
		VALUES('%s','%s','%s','%s','%s','%s','%s', %t);`, args...)
	result, err := db.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	entryId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = entryId
	entry.Running = true
	return nil
}

func (db *Database) GetEntryAsString(id int64) (*EntryDB, error) {
	query := fmt.Sprintf(`SELECT * FROM entries WHERE id = '%d';`, id)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var entryDB EntryDB
	err := db.DB.QueryRowContext(ctx, query).Scan(
		&entryDB.ID,
		&entryDB.Date,
		&entryDB.Begin,
		&entryDB.Finish,
		&entryDB.Hours,
		&entryDB.Project,
		&entryDB.Task,
		&entryDB.Notes,
		&entryDB.Running)
	if err != nil {
		return nil, err
	}
	return &entryDB, nil
}

func (db *Database) GetEntry(id int64) (*Entry, error) {
	query := fmt.Sprintf(`SELECT * FROM entries WHERE id = '%d';`, id)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var entryDB EntryDB
	err := db.DB.QueryRowContext(ctx, query).Scan(
		&entryDB.ID,
		&entryDB.Date,
		&entryDB.Begin,
		&entryDB.Finish,
		&entryDB.Hours,
		&entryDB.Project,
		&entryDB.Task,
		&entryDB.Notes,
		&entryDB.Running)
	if err != nil {
		return nil, err
	}
	entry, err := entryDB.ConvertToEntry()
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (db *Database) UpdateEntry(entry Entry) error {
	args := []any{entry.Date, entry.Begin.String(), entry.Finish.String(), entry.Hours.String(), entry.Project, entry.Task, entry.Notes, entry.Running, entry.ID}
	query := fmt.Sprintf(`UPDATE entries 
				SET date = '%s',
					start = '%s',
					finish = '%s',
					hours = '%s',
					project = '%s',
					task = '%s',
					notes = '%s',
					running = %t
			WHERE id = %d;`, args...)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) AddFinishToEntry(entry Entry) error {
	query := fmt.Sprintf(`UPDATE entries SET finish = '%s', running = false WHERE id = '%d';`, entry.Finish, entry.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteEntry(id int64) error {
	query := fmt.Sprintf(`DELETE FROM entries WHERE id = '%d';`, id)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.DB.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetRunningEntry() (*Entry, error) {
	// We have to make sure that NEVER two entries can be 'running = true'
	query := `SELECT * FROM entries WHERE running = true;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var entryDB EntryDB
	err := db.DB.QueryRowContext(ctx, query).Scan(
		&entryDB.ID,
		&entryDB.Date,
		&entryDB.Begin,
		&entryDB.Finish,
		&entryDB.Hours,
		&entryDB.Project,
		&entryDB.Task,
		&entryDB.Notes,
		&entryDB.Running)
	if err != nil {
		return nil, err
	}
	entry, err := entryDB.ConvertToEntry()
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (db *Database) GetAllEntriesAsString() ([]EntryDB, error) {
	query := `SELECT * FROM entries;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("Got an error reading all entries. Error: %s\n", err.Error())
		return nil, err
	}
	var entries []EntryDB
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entryDB)
	}
	return entries, nil
}

// Gets all entries as 'EntryDB' with times formated AND sorted by 'start ASC'
func (db *Database) GetAllEntriesAsStringWithFormatedTime() ([]EntryDB, error) {
	query := `SELECT * FROM entries ORDER BY start ASC;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("Got an error reading all entries. Error: %s\n", err.Error())
		return nil, err
	}
	var entries []EntryDB
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			fmt.Printf(
				"%s got an error scanning row from db, Error: %s",
				CharError,
				err.Error(),
			)
			return nil, err
		}
		err = entryDB.FormatTimes()
		if err != nil {
			fmt.Printf(
				"%s got an error formatting times in entryDB struct. Struct: %+v, Error: %s",
				CharError,
				entryDB,
				err.Error(),
			)
			return nil, err
		}
		entries = append(entries, entryDB)
	}
	return entries, nil
}

func (db *Database) GetAllEntries() ([]Entry, error) {
	query := `SELECT * FROM entries;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("Got an error reading all entries. Error: %s\n", err.Error())
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			return nil, err
		}
		entry, err := entryDB.ConvertToEntry()
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, nil
}

func (db *Database) GetEntriesViaProject(project string) ([]Entry, error) {
	query := fmt.Sprintf(`SELECT * FROM entries WHERE project = '%s';`, project)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			return nil, err
		}
		entry, err := entryDB.ConvertToEntry()
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, nil
}

func (db *Database) GetEntriesBeforeDate(date time.Time) ([]Entry, error) {
	query := fmt.Sprintf(`SELECT * FROM entries WHERE start < '%s';`, date.String())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			return nil, err
		}
		entry, err := entryDB.ConvertToEntry()
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, nil
}

func (db *Database) GetEntriesAfterDate(date time.Time) ([]Entry, error) {
	query := fmt.Sprintf(`SELECT * FROM entries WHERE start > '%s';`, date.String())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			return nil, err
		}
		entry, err := entryDB.ConvertToEntry()
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, nil
}

// It is possible to filter this for projects
func (db *Database) GetEntriesPerDay(project string) ([]EntriesGroupedByDay, error) {
	query := fmt.Sprintf(`SELECT date, COUNT(DISTINCT(project)), COUNT(DISTINCT(task)), SUM(hours) 
				FROM entries 
				WHERE ((project = '%s') or '%s' = '')
				GROUP BY date;`, project, project)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var EGBY []EntriesGroupedByDay
	for rows.Next() {
		var groupedEntry EntriesGroupedByDay
		var hoursStr string
		err := rows.Scan(
			&groupedEntry.Date,
			&groupedEntry.Projects,
			&groupedEntry.Tasks,
			&hoursStr,
		)
		if err != nil {
			return nil, err
		}
		hoursDec, err := decimal.NewFromString(hoursStr)
		if err != nil {
			fmt.Printf("Could not convert hours from str to decimal. Error: %s", err.Error())
			return nil, err
		}
		groupedEntry.Hours = hoursDec
		EGBY = append(EGBY, groupedEntry)
	}
	return EGBY, nil
}

func (db *Database) GetUniqueProjects() ([]string, error) {
	query := `SELECT DISTINCT(project) FROM entries;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var projects []string
	for rows.Next() {
		var project string
		err := rows.Scan(&project)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}
func createDefaultTables(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS entries(
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			date  TEXT NOT NULL,
			start TEXT NOT NULL,
			finish TEXT,
			hours  FLOAT,
			project TEXT NOT NULL,
			task   TEXT NOT NULL,
			notes  TEXT,
			running BOOL);`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		fmt.Printf("error while making default table. Error: %s", err.Error())
		return err
	}
	return nil
}

// Will give summed hours per date, irrgeardless of project
// Returns 'date | SUM(hours)'
func (db *Database) GetHoursTrackedPerDay(cf calendarTimeFrame) ([]GroupedEntries, error) {
	query := fmt.Sprintf(`SELECT date, SUM(hours) FROM entries 
						WHERE start >= '%s' 
						AND start <= '%s'
						GROUP BY date
						ORDER BY START ASC;`, cf.since, cf.until)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("ran into an error when reading the database. Error: %s", err.Error())
		return nil, err
	}
	var dailyEntries []GroupedEntries
	for rows.Next() {
		var dailyEntry GroupedEntries
		var hoursStr string
		err := rows.Scan(
			&dailyEntry.Date,
			&hoursStr,
		)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		hoursDec, err := decimal.NewFromString(hoursStr)
		if err != nil {
			fmt.Printf("Could not convert hours from str to decimal. Error: %s", err.Error())
			return nil, err
		}
		dailyEntry.Hours = hoursDec
		dailyEntries = append(dailyEntries, dailyEntry)
	}
	return dailyEntries, nil
}

func (db *Database) GetHoursTrackedPerDayAndProject(cf calendarTimeFrame) ([]GroupedProjectHours, error) {
	query := fmt.Sprintf(`SELECT date, project SUM(hours) FROM entries 
						WHERE start < '%s' 
						AND start > '%s'
						GROUP BY date, project
						ORDER BY START ASC;`, cf.since, cf.until)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var dailyEntries []GroupedProjectHours
	for rows.Next() {
		var dailyProjectEntry GroupedProjectHours
		var hoursStr string
		err := rows.Scan(
			&dailyProjectEntry.Date,
			&dailyProjectEntry.Project,
			&hoursStr,
		)
		if err != nil {
			return nil, err
		}
		hoursDec, err := decimal.NewFromString(hoursStr)
		if err != nil {
			fmt.Printf("Could not convert hours from str to decimal. Error: %s", err.Error())
			return nil, err
		}
		dailyProjectEntry.Hours = hoursDec
		dailyEntries = append(dailyEntries, dailyProjectEntry)
	}
	return dailyEntries, nil
}

// This gives you back all horus worked accumulated, grouped per month.
func (db *Database) GetHoursTrackerPerMonth(cf calendarTimeFrame) ([]GroupedEntries, error) {
	query := `SELECT strftime('%m', start) as month, SUM(hours) FROM entries `
	query += fmt.Sprintf(`WHERE start >= '%s' AND start <= '%s' `, cf.since, cf.until)
	query += `GROUP BY strftime('%m', start) ORDER BY START ASC;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("ran into an error when reading the database. Error: %s", err.Error())
		return nil, err
	}
	var monthlyEntries []GroupedEntries
	for rows.Next() {
		var monthlyEntry GroupedEntries
		var hoursStr string
		err := rows.Scan(
			&monthlyEntry.Date,
			&hoursStr,
		)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		hoursDec, err := decimal.NewFromString(hoursStr)
		if err != nil {
			fmt.Printf("Could not convert hours from str to decimal. Error: %s", err.Error())
			return nil, err
		}
		monthlyEntry.Hours = hoursDec
		monthlyEntries = append(monthlyEntries, monthlyEntry)
	}
	return monthlyEntries, nil
}


// This gives you back all horus worked accumulated, grouped per month.
func (db *Database) GetHoursTrackerPerWeek(cf calendarTimeFrame) ([]GroupedEntries, error) {
	query := `SELECT strftime('%W', start) as date, SUM(hours) FROM entries `
	query += fmt.Sprintf(`WHERE start >= '%s' AND start <= '%s' `, cf.since, cf.until)
	query += `GROUP BY strftime('%W', start) ORDER BY START ASC;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Printf("ran into an error when reading the database. Error: %s", err.Error())
		return nil, err
	}
	var weeklyEntries []GroupedEntries
	for rows.Next() {
		var weeklyEntry GroupedEntries
		var hoursStr string
		err := rows.Scan(
			&weeklyEntry.Date,
			&hoursStr,
		)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		hoursDec, err := decimal.NewFromString(hoursStr)
		if err != nil {
			fmt.Printf("Could not convert hours from str to decimal. Error: %s", err.Error())
			return nil, err
		}
		weeklyEntry.Hours = hoursDec
		weeklyEntries = append(weeklyEntries, weeklyEntry)
	}
	return weeklyEntries, nil
}



// Unsure if 'max' or 'sum' of hours even works as it's a 'string' type in the DB
// TODO: Test
func (db *Database) GetLongestEntryPerDay(cf calendarTimeFrame) ([]Entry, error) {
	query := fmt.Sprintf(`SELECT *, MAX(hours), FROM entries 
						WHERE start < '%s' 
						AND start > '%s'
						GROUP BY date 
						ORDER BY START ASC;`, cf.since, cf.until)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entryDB EntryDB
		err := rows.Scan(
			&entryDB.ID,
			&entryDB.Date,
			&entryDB.Begin,
			&entryDB.Finish,
			&entryDB.Hours,
			&entryDB.Project,
			&entryDB.Task,
			&entryDB.Notes,
			&entryDB.Running)
		if err != nil {
			return nil, err
		}
		entry, err := entryDB.ConvertToEntry()
		if err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return entries, nil
}
