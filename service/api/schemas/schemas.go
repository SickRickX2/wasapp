package schemas

import "time"

type User struct {
	ID        string    `json:"userId"`
	Name      string    `json:"name"`
	UserName  *string   `json:"userName,omitempty"`
	PfpURL    *string   `json:"pfp,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
