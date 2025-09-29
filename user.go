package classosbackend

import "errors"

type User struct {
	ID        int     `json:"id" db:"id"`
	Name      string  `json:"name" db:"name" binding:"required"`
	Username  string  `json:"username" db:"username" binding:"required"`
	Password  string  `json:"password" db:"password_hash" binding:"required"`
	Role      string  `json:"role" db:"role"`
	GroupID   *int    `json:"group_id,omitempty" db:"group_id"`
	GroupName *string `json:"group_name" db:"group_name"`
}

type UpdateUserInput struct {
	Name      *string `json:"name"`
	Username  *string `json:"username"`
	Password  *string `json:"password"`
	Role      *string `json:"role"`
	GroupID   *int    `json:"group_id"`
	GroupName *string `json:"group_name"`
}

func (i UpdateUserInput) Validate() error {
	if i.Name == nil && i.Username == nil && i.Role == nil && i.GroupID == nil {
		return errors.New("update structure has no values")
	}

	return nil
}
