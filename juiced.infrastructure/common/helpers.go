package common

import (
	"math/rand"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kirsle/configdir"
	_ "github.com/mattn/go-sqlite3"
)

// Removes the string x from the slice s
func RemoveFromSlice(s []string, x string) []string {
	var position int
	for i, r := range s {
		if r == x {
			position = i
		}
	}

	s[position] = s[len(s)-1]

	return s[:len(s)-1]
}

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var runes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

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
	}

	return err
}

// GetDatabase retrieves the database connection
func GetDatabase() *sqlx.DB {
	return database
}
