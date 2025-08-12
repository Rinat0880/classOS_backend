package classosbackend

import "time"

type Group struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WhitelistEntry struct {
    ID        int64     `json:"id"`
    GroupID   int64     `json:"group_id"`
    Value     string    `json:"value"` // Например, IP или email
    CreatedAt time.Time `json:"created_at"`
}

type Settings struct {
    ID        int64     `json:"id"`
    Key       string    `json:"key"`
    Value     string    `json:"value"`
    UpdatedAt time.Time `json:"updated_at"`
}