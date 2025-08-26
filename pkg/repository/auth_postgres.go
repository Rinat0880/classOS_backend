package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	classosbackend "github.com/rinat0880/classOS_backend"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user classosbackend.User) (int, error) {
	var id int
	query := fmt.Sprintf("INSERT INTO %s (name, username, role, password_hash) values ($1, $2, 'client', $3) RETURNING id", usersTable)
	row := r.db.QueryRow(query, user.Name, user.Username, user.Password)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthPostgres) GetUser(username, password string) (classosbackend.User, error) {
	var user classosbackend.User
	query := fmt.Sprintf("SELECT id, role FROM %s WHERE username=$1 AND password_hash=$2", usersTable)
	err := r.db.Get(&user, query, username, password)
	
	fmt.Printf("Retrieved user: ID=%d, Role=%s\n", user.ID, user.Role) 
	
	return user, err
}
