package datadome

type DatadomeInfo struct {
	InitialCID string `json:"cid"`
	CID        string `json:"-"`
	Hash       string `json:"hsh"`
	T          string `json:"t"`
	S          int    `json:"s"`
	Host       string `json:"host"`
}
