package common

import (
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var runes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var schema = `
	CREATE TABLE IF NOT EXISTS userInfo (
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

// ConnectToDatabase returns a database object
func ConnectToDatabase() (*sqlx.DB, error) {
	database, err := sqlx.Connect("sqlite3", "./juiced.db")
	if err != nil {
		return database, err
	}

	_, err = database.Exec(schema)
	return database, err
}
