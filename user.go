package classosbackend

type User struct {
	ID       int   `json:"-" db:"id"`
	Name     string `json:"name" db:"name" binding:"required"`
	Username string `json:"username" db:"username" binding:"required"`
	Password string `json:"password" db:"password_hash" binding:"required"` 
	Role     string `json:"role" db:"role"`
}
