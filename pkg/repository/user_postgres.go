package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	classosbackend "github.com/rinat0880/classOS_backend"
)

type UserPostgres struct {
	db *sqlx.DB
}

func NewUserPostgres(db *sqlx.DB) *UserPostgres {
	return &UserPostgres{db: db}
}

func (r *UserPostgres) Create(groupId int, user classosbackend.User) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var user_id int
	createUserQuery := fmt.Sprintf("INSERT INTO %s (name, username, role, password_hash) values ($1, $2, $3, $4) RETURNING id", usersTable)
	row := tx.QueryRow(createUserQuery, user.Name, user.Username, user.Role, user.Password)
	err = row.Scan(&user_id)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	createUserListsQuery := fmt.Sprintf("INSERT INTO %s (group_id, user_id) values ($1, $2)", users_listsTable)
	_, err = tx.Exec(createUserListsQuery, groupId, user_id)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return user_id, tx.Commit()
}


func (r *UserPostgres) GetAll(userId, groupId int) ([]classosbackend.User, error) {
	var users []classosbackend.User
	query := fmt.Sprintf("SELECT u.id, u.name, u.username, u.role, u.password_hash FROM %s u JOIN %s ul ON u.id = ul.user_id WHERE ul.group_id = $1", usersTable, users_listsTable)
	if err := r.db.Select(&users, query, groupId); err != nil {
		return nil, err
	}

	return users, nil
}
