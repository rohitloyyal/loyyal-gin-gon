package models

import "time"

/***********
Sample Identity Object
	{
	    "username": "0xea900098147c235E1C3", //system generated
	    "password": "password", // hashed before storing
	    "identityType": "partner", // operator, partner, consumer
	    "personalDetails": {
	        "firstName": "Rohit", //required
	        "lastName": "Roy", //required
	        "emailAddress": "rohit@airlines.ae", // required
	        "countryCode": 971, // required
	        "mobileNo": 535433322, // required
	        "entityName": "Test Airlines",
	        "locality": "JLT", // optional
	        "city": "Dubai", // optional
	        "country": "United Arab Emirates",
	        "zipcode": "535433" // optional
	    },
	    "wallets": [
	        {
	            "id": "0x5E1C3", // system generated
	            "name": "wallet1",
	            "metadata": {
	                // details captured from external systems
	            },
	            "assets": [
	                {
	                    "currency": "points",
	                    "balance": 4333
	                }
	            ],
	            "walletType": "default", // default (in/out, all will have one default wallet on identity creation),
	            // burn ( only in, will be available for the partner only
	            "status": "active" // active, disabled
	        }
	    ],
	    "createdAt": "09-03-2022", // system generated
	    "creator": "/user/admin",
	    "channel": "loyyalchannel",
	    "status": "active", // created, verification pending, active, disabled
	    "lastUpdatedOn": "13-12-2022",
	    "lastLoggedInOn": "13-12-2022",
	    "lastPasswordResetOn": "03-12-2022"
	}
*
*/

type PersonalDetails struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
	CountryCode  int    `json:"countryCode"`
	MobileNo     int    `json:"mobileNo"`
	Locality     string `json:"locality"`
	City         string `json:"city"`
	Country      string `json:"country"`
	ZipCode      string `json:"zipcode"`
}

type WalletRef struct {
	Ref string `json:"ref"`
}

type Identity struct {
	DocType             string          `json:"docType"`
	Identifier          string          `json:"identifier"`
	Username            string          `json:"username"`
	Password            string          `json:"password"`
	IdentityType        string          `json:"identityType"`
	PersonalDetails     PersonalDetails `json:"personalDetails"`
	EntityName          string          `json:"entityName"`
	Wallets             []WalletRef     `json:"wallets"`
	Creator             string          `json:"creator"`
	Channel             string          `json:"channel"`
	Status              string          `json:"status"`
	CreatedAt           time.Time       `json:"createdAt"`
	LastUpdatedAt       time.Time       `json:"lastUpdatedAt"`
	LastUpdatedBy       string          `json:"lastUpdatedBy"`
	LastLoggedInAt      time.Time       `json:"lastLoggedInAt"`
	LastPasswordResetOn time.Time       `json:"lastPasswordResetOn"`
	Hash                string          `json:"hashed"`
	IsDeleted           bool            `json:"isDeleted"`
}
