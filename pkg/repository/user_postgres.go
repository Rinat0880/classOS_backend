package repository

import (
	"fmt"
	"strings"

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

func (r *UserPostgres) GetAll(checkerId, groupId int) ([]classosbackend.User, error) {
	var users []classosbackend.User
	query := fmt.Sprintf("SELECT u.id, u.name, u.username, u.role, u.password_hash FROM %s u JOIN %s ul ON u.id = ul.user_id WHERE ul.group_id = $1", usersTable, users_listsTable)
	if err := r.db.Select(&users, query, groupId); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserPostgres) GetById(checkerId, user_id int) (classosbackend.User, error) {
	var user classosbackend.User
	query := fmt.Sprintf(`
		SELECT u.id, u.name, u.username, u.role,u.password_hash, ul.group_id, g.name as group_name FROM %s u LEFT JOIN %s ul ON u.id = ul.user_id LEFT JOIN %s g ON ul.group_id = g.id WHERE u.id = $1`, usersTable, users_listsTable, groupsTable)

	if err := r.db.Get(&user, query, user_id); err != nil {
		return user, err
	}

	return user, nil
}

func (r *UserPostgres) Delete(checkerId, user_id int) error {
	query := fmt.Sprintf(`Delete FROM %s WHERE id = $1`, usersTable)

	_, err := r.db.Exec(query, user_id)

	return err
}

func (r *UserPostgres) Update(checkerId, user_id int, input classosbackend.UpdateUserInput) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	userSetValues := make([]string, 0)
	userArgs := make([]interface{}, 0)
	argId := 1

	if input.Name != nil {
		userSetValues = append(userSetValues, fmt.Sprintf("name=$%d", argId))
		userArgs = append(userArgs, *input.Name)
		argId++
	}

	if input.Username != nil {
		userSetValues = append(userSetValues, fmt.Sprintf("username=$%d", argId))
		userArgs = append(userArgs, *input.Username)
		argId++
	}

	if input.Password != nil {
		userSetValues = append(userSetValues, fmt.Sprintf("password_hash=$%d", argId))
		userArgs = append(userArgs, *input.Password)
		argId++
	}

	if input.Role != nil {
		userSetValues = append(userSetValues, fmt.Sprintf("role=$%d", argId))
		userArgs = append(userArgs, *input.Role)
		argId++
	}

	if len(userSetValues) > 0 {
		setQuery := strings.Join(userSetValues, ", ")
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", usersTable, setQuery, argId)
		userArgs = append(userArgs, user_id)

		_, err = tx.Exec(query, userArgs...)
		if err != nil {
			return err
		}
	}

	if input.GroupID != nil {
		query := fmt.Sprintf(`
			INSERT INTO %s (user_id, group_id) 
			VALUES ($1, $2)
			ON CONFLICT (user_id, group_id) DO NOTHING
		`, users_listsTable)
		
		_, err = tx.Exec(query, user_id, *input.GroupID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}