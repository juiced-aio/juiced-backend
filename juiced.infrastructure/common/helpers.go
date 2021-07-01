package common

import (
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"github.com/jmoiron/sqlx"
	"github.com/kirsle/configdir"
	_ "github.com/mattn/go-sqlite3"
)

// Returns true if it finds the string x in the slice s
func InSlice(s []string, x string) bool {
	for _, i := range s {
		if i == x {
			return true
		}
	}
	return false
}

// Removes the string x from the slice s
func RemoveFromSlice(s []string, x string) []string {
	if !InSlice(s, x) {
		return s
	}
	var position int
	for i, r := range s {
		if r == x {
			position = i
		}
		return append(s[:position], s[position+1:]...)
	}

	return s
}

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var runes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func FindInString(str string, start string, end string) (string, error) {
	comp := regexp.MustCompile(fmt.Sprintf("%v(.*?)%v", start, end))
	comp.MatchString(str)

	o := comp.FindStringSubmatch(str)
	if len(o) == 0 {
		return "", errors.New("string not found")
	}
	parsed := o[1]
	if parsed == "" {
		return parsed, errors.New("string not found")
	}

	return parsed, nil

}

// RandID returns a random n-digit ID of digits and uppercase letters
func RandID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[seededRand.Intn(len(runes))]
	}
	return string(b)
}

var database *sqlx.DB

// InitDatabase initializes the database singleton
func InitDatabase() error {
	var err error

	configPath := configdir.LocalConfig("juiced")
	err = configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	filename := filepath.Join(configPath, "juiced.db")
	database, err = sqlx.Connect("sqlite3", filename)
	if err != nil {
		return err
	}

	for _, schema := range schemas {
		_, err = database.Exec(schema)
		tableName, _ := FindInString(schema, "EXISTS ", " \\(")
		missing, extra := CompareColumns(ParseColumns(schema), GetCurrentColumns(schema))
		for i := range extra {
			extraSplit := strings.Split(extra[i], "|")
			_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v DROP COLUMN %v", tableName, extraSplit[0]))
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

		for i := range missing {
			missingSplit := strings.Split(missing[i], "|")
			_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v DEFAULT ''", tableName, missingSplit[0], missingSplit[1]))
			if err != nil {
				fmt.Println(err)
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

func ProxyCleaner(proxyDirty entities.Proxy) string {
	if proxyDirty.Host == "" {
		return ""
	}
	if proxyDirty.Username == "" && proxyDirty.Password == "" {
		return fmt.Sprintf("http://%s:%s", proxyDirty.Host, proxyDirty.Port)
	} else {
		return fmt.Sprintf("http://%s:%s@%s:%s", proxyDirty.Username, proxyDirty.Password, proxyDirty.Host, proxyDirty.Port)
	}
}

func ParseColumns(schema string) (columnNames []string) {
	schema = strings.ReplaceAll(schema, "\n", "")
	schema = strings.ReplaceAll(schema, "\t", "")
	inside, _ := FindInString(schema, "\\(", "\\)")
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
	tableName, _ := FindInString(schema, "EXISTS ", " \\(")
	rows, _ := database.Queryx("PRAGMA table_info(" + tableName + ");")

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
