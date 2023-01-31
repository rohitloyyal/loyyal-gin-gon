package models

type User struct {
	DocType  string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	
}
