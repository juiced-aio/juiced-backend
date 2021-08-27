package database

type DatabaseNotInitializedError struct{}

func (e *DatabaseNotInitializedError) Error() string {
	return "database not initialized"
}
