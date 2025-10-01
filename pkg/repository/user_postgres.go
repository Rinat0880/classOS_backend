package repository

import (
	"database/sql"
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

func (r *UserPostgres) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *UserPostgres) CreateWithTx(tx *sql.Tx, groupId int, user classosbackend.User) (int, error) {
	var userId int
	createUserQuery := fmt.Sprintf("INSERT INTO %s (name, username, role, password_hash) values ($1, $2, $3, $4) RETURNING id", usersTable)
	row := tx.QueryRow(createUserQuery, user.Name, user.Username, user.Role, user.Password)
	err := row.Scan(&userId)
	if err != nil {
		return 0, err
	}

	createUserListsQuery := fmt.Sprintf("INSERT INTO %s (group_id, user_id) values ($1, $2)", users_listsTable)
	_, err = tx.Exec(createUserListsQuery, groupId, userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (r *UserPostgres) UpdateWithTx(tx *sql.Tx, checkerId, userId int, input classosbackend.UpdateUserInput) error {
	userSetValues := make([]string, 0)
	userArgs := make([]interface{}, 0)
	argId := 1

	if input.Name != nil {
		userSetValues = append(userSetValues, fmt.Sprintf("name=$%d", argId))
		userArgs = append(userArgs, *input.Name)
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
		userArgs = append(userArgs, userId)

		_, err := tx.Exec(query, userArgs...)
		if err != nil {
			return err
		}
	}

	if input.GroupID != nil {
	    deleteQuery := fmt.Sprintf(`
	        DELETE FROM %s  
	        WHERE user_id = $1
	    `, users_listsTable)
	
	    _, err := tx.Exec(deleteQuery, userId)
	    if err != nil {
	        return err
	    }

		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (user_id, group_id)
			VALUES ($1, $2)
		`, users_listsTable)
		_, err = tx.Exec(insertQuery, userId, *input.GroupID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *UserPostgres) DeleteWithTx(tx *sql.Tx, checkerId, userId int) error {
	query := fmt.Sprintf(`Delete FROM %s WHERE id = $1`, usersTable)
	_, err := tx.Exec(query, userId)
	return err
}

func (r *UserPostgres) Create(groupId int, user classosbackend.User) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var userId int
	createUserQuery := fmt.Sprintf("INSERT INTO %s (name, username, role, password_hash) values ($1, $2, $3, $4) RETURNING id", usersTable)
	row := tx.QueryRow(createUserQuery, user.Name, user.Username, user.Role, user.Password)
	err = row.Scan(&userId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	createUserListsQuery := fmt.Sprintf("INSERT INTO %s (group_id, user_id) values ($1, $2)", users_listsTable)
	_, err = tx.Exec(createUserListsQuery, groupId, userId)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return userId, tx.Commit()
}

func (r *UserPostgres) GetAll(checkerId int) ([]classosbackend.User, error) {
	var users []classosbackend.User
	query := fmt.Sprintf(`
		SELECT u.id, u.name, u.username, u.role, 
			   COALESCE(ul.group_id, 0) as group_id, 
			   COALESCE(g.name, '') as group_name 
		FROM %s u 
		LEFT JOIN %s ul ON u.id = ul.user_id 
		LEFT JOIN %s g ON ul.group_id = g.id 
		WHERE u.id != 1`, 
		usersTable, users_listsTable, groupsTable)
	
	err := r.db.Select(&users, query)
	return users, err
}

func (r *UserPostgres) GetById(checkerId, userId int) (classosbackend.User, error) {
	var user classosbackend.User
	query := fmt.Sprintf(`
		SELECT u.id, u.name, u.username, u.role, ul.group_id, g.name as group_name 
		FROM %s u 
		LEFT JOIN %s ul ON u.id = ul.user_id 
		LEFT JOIN %s g ON ul.group_id = g.id 
		WHERE u.id = $1`, usersTable, users_listsTable, groupsTable)

	err := r.db.Get(&user, query, userId)
	return user, err
}

func (r *UserPostgres) Delete(checkerId, userId int) error {
	query := fmt.Sprintf(`Delete FROM %s WHERE id = $1`, usersTable)
	_, err := r.db.Exec(query, userId)
	return err
}

func (r *UserPostgres) Update(checkerId, userId int, input classosbackend.UpdateUserInput) error {
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
		userArgs = append(userArgs, userId)

		_, err = tx.Exec(query, userArgs...)
		if err != nil {
			return err
		}
	}

	if input.GroupID != nil {
	    deleteQuery := fmt.Sprintf(`
	        DELETE FROM %s  
	        WHERE user_id = $1
	    `, users_listsTable)
	
	    _, err := tx.Exec(deleteQuery, userId)
	    if err != nil {
	        return err
	    }

		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (user_id, group_id)
			VALUES ($1, $2)
		`, users_listsTable)
		_, err = tx.Exec(insertQuery, userId, *input.GroupID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}