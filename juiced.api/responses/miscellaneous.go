package responses

type MiscellaneousResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
}
