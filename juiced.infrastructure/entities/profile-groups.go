package entities

// ProfileGroup is a class that holds a list of Profile IDs
type ProfileGroup struct {
	GroupID          string     `json:"groupID" db:"groupID"`
	Name             string     `json:"name" db:"name"`
	ProfileIDs       []string   `json:"profileIDs"`
	ProfileIDsJoined string     `json:"-" db:"profileIDsJoined"`
	CreationDate     int64      `json:"creationDate" db:"creationDate"`
	Profiles         []*Profile `json:"profiles"`
}
