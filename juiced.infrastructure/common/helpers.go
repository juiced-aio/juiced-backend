package common

import (
	"math/rand"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kirsle/configdir"
	_ "github.com/mattn/go-sqlite3"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var runes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var schema = `
	CREATE TABLE IF NOT EXISTS userInfo (
		ID INTEGER,
		email TEXT,
		licenseKey TEXT,
		deviceName TEXT,
		discordID TEXT,
		discordUsername TEXT,
		discordAvatarURL TEXT,
		activationToken TEXT,
		refreshToken TEXT,
		expiresAt INTEGER
	)
`

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
	configPath := configdir.LocalConfig("Juiced")
	err = configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	filename := filepath.Join(configPath, "juiced.db")
	database, err = sqlx.Connect("sqlite3", filename)
	if err != nil {
		return err
	}

	_, err = database.Exec(schema)
	return err
}

// GetDatabase retrieves the database connection
func GetDatabase() *sqlx.DB {
	return database
}
