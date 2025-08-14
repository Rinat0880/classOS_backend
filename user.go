package classosbackend

type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}
