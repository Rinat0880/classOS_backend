package classosbackend

import "time"

type User struct {
	ID        uint      `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`    
	Role      string    `json:"role"` 
	CreatedAt time.Time `json:"created_at"`
}
