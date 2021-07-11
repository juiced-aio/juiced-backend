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
var runesWithLower = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz123456789")

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

// RandString returns a random n-digit string of digits and uppercase/lowercase letters
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runesWithLower[seededRand.Intn(len(runesWithLower))]
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
		if err != nil {
			fmt.Println(err)
		}
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
			if strings.Contains(missing[i], "TEXT") {
				_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v DEFAULT ''", tableName, missingSplit[0], missingSplit[1]))
			} else if strings.Contains(missing[i], "INTEGER") {
				_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v DEFAULT 0", tableName, missingSplit[0], missingSplit[1]))
			} else {
				_, err = database.Exec(fmt.Sprintf("ALTER TABLE %v ADD COLUMN %v %v", tableName, missingSplit[0], missingSplit[1]))
			}
			if err != nil {
				fmt.Println(err)
				return err
			}
		}

	}
	_, err = database.Exec("DELETE FROM checkouts WHERE time < 1625782953")
	if err != nil {
		fmt.Println(err)
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

func DetectCardType(cardNumber []byte) string {
	matched, _ := regexp.Match(`^4`, cardNumber)
	if matched {
		return "Visa"
	}

	matched, _ = regexp.Match(`^(5[1-5][0-9]{14}|2(22[1-9][0-9]{12}|2[3-9][0-9]{13}|[3-6][0-9]{14}|7[0-1][0-9]{13}|720[0-9]{12}))$`, cardNumber)
	if matched {
		return "Mastercard"
	}

	matched, _ = regexp.Match(`^3[47]`, cardNumber)
	if matched {
		return "AMEX"
	}

	matched, _ = regexp.Match(`^(6011|622(12[6-9]|1[3-9][0-9]|[2-8][0-9]{2}|9[0-1][0-9]|92[0-5]|64[4-9])|65)`, cardNumber)
	if matched {
		return "Discover"
	}

	matched, _ = regexp.Match(`^36`, cardNumber)
	if matched {
		return "Diners"
	}

	matched, _ = regexp.Match(`^30[0-5]`, cardNumber)
	if matched {
		return "Diners - Carte Blanche"
	}

	matched, _ = regexp.Match(`^35(2[89]|[3-8][0-9])`, cardNumber)
	if matched {
		return "JCB"
	}

	matched, _ = regexp.Match(`^(4026|417500|4508|4844|491(3|7))`, cardNumber)
	if matched {
		return "Visa Electron"
	}

	return ""
}
