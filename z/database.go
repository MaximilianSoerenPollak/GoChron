package z

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
		return nil, err
	}
	err = createDefaultTables(db)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

func (db *Database) AddEntry(entry *Entry) error {
	query := `INSERT INTO entries(date, start, finish, hours, project, task, notes) 
				VALUES ('?','?','?','?','?','?','?', true);`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := db.DB.ExecContext(ctx, query, entry)
	if err != nil {
		return err
	}
	entryId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = entryId
	return nil
}

func (db *Database) GetEntry(id int64) (*Entry, error) {
	query := `SELECT * FROM entries WHERE id = '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var entry *Entry
	err := db.DB.QueryRowContext(ctx, query, id).Scan(&entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (db *Database) UpdateEntry(entry Entry) error {
	query := `UPDATE entries 
				SET date = '?',
				SET start = '?',
				SET finish = '?',
				SET hours = '?',
				SET project = '?',
				SET task = '?',
				SET notes = '?',
				SET running = '?'
			WHERE id = '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{entry.Date, entry.Begin, entry.Finish, entry.Hours, entry.Project, entry.Task, entry.Notes, entry.ID}
	_, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) AddFinishToEntry(entry Entry) error {
	query := `UPDATE entries SET finish = '?', SET running = false, WHERE id = '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{entry.Finish, entry.ID}
	_, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteEntry(id int64) error {
	query := `DELETE FROM entries WHERE id = '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.DB.ExecContext(ctx, query, id)
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
	var entry *Entry
	err := db.DB.QueryRowContext(ctx, query).Scan(&entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (db *Database) GetAllEntries() ([]Entry, error) {
	query := `SELECT * FROM entries;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(&entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (db *Database) GetEntriesViaProject(project string) ([]Entry, error) {
	query := `SELECT * FROM entries WHERE project = '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query, project)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(&entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (db *Database) GetEntriesBeforeDate(date time.Time) ([]Entry, error) {
	query := `SELECT * FROM entries WHERE start < '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query, project)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(&entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (db *Database) GetEntriesAfterDate(date time.Time) ([]Entry, error) {
	query := `SELECT * FROM entries WHERE start > '?';`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query, project)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(&entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// It is possible to filter this for projects
func (db *Database) GetEntriesPerDay(project string) ([]EntriesGroupedByDay, error) {
	query := `SELECT date, COUNT(DISTINCT(project)), COUNT(DISTINCT(task)), SUM(hours) 
				FROM entries 
				WHERE ((project = '?') or '?' = '')
				GROUP BY date;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query, project)
	if err != nil {
		rows.Close()
		return nil, err
	}
	var EGBY []EntriesGroupedByDay
	for rows.Next() {
		var groupedEntry EntriesGroupedByDay
		err := rows.Scan(&groupedEntry)
		if err != nil {
			return nil, err
		}
		EGBY = append(EGBY, groupedEntry)
	}
	return EGBY, nil
}

func (db *Database) GetUniqueProjects() ([]string, error) {
	query := `SELECT DISTINCT(project) FROM entries;`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := db.DB.QueryContext(ctx, query, project)
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
			date  NOT NULL,
			start TIMESTAMP/DATETIME NOT NULL,
			finish TIMESTAMP/DATETIME,
			hours  FLOAT,
			project TEXT NOT NULL,
			task   TEXT NOT NULL,
			notes  TEXT
			running BOOL);`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return nil
}
