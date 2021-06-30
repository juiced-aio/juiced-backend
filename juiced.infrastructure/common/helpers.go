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
		for i := range missing {
			missingSplit := strings.Split(missing[i], "|")
			_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v", tableName, missingSplit[0], missingSplit[1]))
			if err != nil {
				return err
			}
		}
		for i := range extra {
			_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v DROP COLUMN %v", tableName, extra[i]))
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
		columnNames = append(columnNames, column[1].(string))
	}
	return
}

func CompareColumns(x []string, y []string) ([]string, []string) {
	var missing []string
	var extra []string
	for i := range x {
		var inside bool
		for _, name := range y {
			if name == strings.Split(x[i], "|")[0] {
				inside = true
			}
		}
		if !inside {
			missing = append(missing, x[i])
		}
	}
	for i := range y {
		var inside bool
		for _, name := range x {
			if strings.Split(name, "|")[0] == y[i] {
				inside = true
			}
		}
		if !inside {
			extra = append(extra, y[i])
		}
	}

	return missing, extra
}
