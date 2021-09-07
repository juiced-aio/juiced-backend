package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"backend.juicedbot.io/juiced.infrastructure/util"
	"github.com/jmoiron/sqlx"
	"github.com/kirsle/configdir"
	_ "github.com/mattn/go-sqlite3"
)

var database *sqlx.DB

const DB_FILENAME = "juiced3.db"
const DB_TEST_FILENAME = "juiced_test.db"

func InitDatabase() error {
	filepath, err := GetDBFilePath(DB_FILENAME)
	if err != nil {
		return err
	}
	return InitDatabaseHelper(filepath)
}

func InitTestDatabase() error {
	filepath, err := GetDBFilePath(DB_TEST_FILENAME)
	if err != nil {
		return err
	}
	if err := DeleteDatabaseFile(filepath); err != nil {
		return err
	}
	if _, err := os.Create(filepath); err != nil {
		return err
	}
	return InitDatabaseHelper(filepath)
}

func GetDBFilePath(filename string) (string, error) {
	configPath := configdir.LocalConfig("juiced")
	err := configdir.MakePath(configPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, filename), nil
}

func DeleteDatabaseFile(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(filepath)
}

func InitDatabaseHelper(filepath string) error {
	var err error
	database, err = sqlx.Connect("sqlite3", "file:"+filepath+"?cache=shared&mode=rwc")
	if err != nil {
		return err
	}

	for _, schema := range schemas {
		_, err = database.Exec(schema)
		if err != nil {
			return err
		}
		tableName, _ := util.FindInString(schema, "EXISTS ", " \\(")
		missing, extra := CompareColumns(ParseColumns(schema), GetCurrentColumns(schema))
		for i := range extra {
			extraSplit := strings.Split(extra[i], "|")
			_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v DROP COLUMN %v", tableName, extraSplit[0]))
			if err != nil {
				return err
			}
		}

		for i := range missing {
			missingSplit := strings.Split(missing[i], "|")
			if strings.Contains(missing[i], "TEXT") {
				_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v DEFAULT ''", tableName, missingSplit[0], missingSplit[1]))
			} else if strings.Contains(missing[i], "INTEGER") {
				_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v DEFAULT 0", tableName, missingSplit[0], missingSplit[1]))
			} else {
				_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v", tableName, missingSplit[0], missingSplit[1]))
			}
			if err != nil {
				return err
			}
		}
	}

	return err
}

// GetDatabase retrieves the database connection
func GetDatabase() *sqlx.DB {
	return database
}

func ParseColumns(schema string) (columnNames []string) {
	schema = strings.ReplaceAll(schema, "\n", "")
	schema = strings.ReplaceAll(schema, "\t", "")
	inside, _ := util.FindInString(schema, "\\(", "\\)")
	columns := strings.Split(inside, ",")
	for _, column := range columns {
		if strings.Contains(column, " ") {
			columnSplit := strings.Split(column, " ")
			columnNames = append(columnNames, columnSplit[0]+"|"+columnSplit[1])
		}
	}
	return
}

func GetCurrentColumns(schema string) (columnNames []string) {
	tableName, _ := util.FindInString(schema, "EXISTS ", " \\(")
	rows, _ := database.Queryx("PRAGMA table_info(" + tableName + ");")

	defer rows.Close()
	for rows.Next() {
		column, _ := rows.SliceScan()
		columnNames = append(columnNames, column[1].(string)+"|"+column[2].(string))
	}
	return
}

func CompareColumns(x []string, y []string) ([]string, []string) {
	var missing []string
	var extra []string

	for _, extraColumn1 := range x {
		var inside bool
		extraColumnSplit1 := strings.Split(extraColumn1, "|")
		for _, extraColumn2 := range y {
			extraColumnSplit2 := strings.Split(extraColumn2, "|")
			if extraColumnSplit2[0] == extraColumnSplit1[0] && extraColumnSplit2[1] == extraColumnSplit1[1] {
				inside = true
			}
		}
		if !inside {
			missing = append(missing, extraColumn1)
		}
	}

	for _, missingColumn1 := range y {
		var inside bool
		missingColumnSplit1 := strings.Split(missingColumn1, "|")
		for _, missingColumn2 := range x {

			missingColumnSplit2 := strings.Split(missingColumn2, "|")
			if missingColumnSplit2[0] == missingColumnSplit1[0] && missingColumnSplit2[1] == missingColumnSplit1[1] {
				inside = true
			}
		}
		if !inside {
			extra = append(extra, missingColumn1)
		}
	}

	return missing, extra
}
