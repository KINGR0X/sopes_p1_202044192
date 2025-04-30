package structs

type Tweet struct {
	Description string `json:"description"`
	Country     string `json:"country"`
	Weather     int `json:"weather"`
}
