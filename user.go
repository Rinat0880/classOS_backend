package classosbackend


type User struct {
	ID        uint      `json:"id"`
	FullName  string    `json:"full_name"`
	Password  string    `json:"-"`    
	Role      string    `json:"role"` 
}
