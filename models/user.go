package models

type User struct {
	DocType  string `json:"docType"`
	Username string `json:"username"`
	Password string `json:"password"`
	
}
