package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	classosbackend "github.com/rinat0880/classOS_backend"
)

type GroupPostgres struct {
	db *sqlx.DB
}

func NewGroupPostgres(db *sqlx.DB) *GroupPostgres {
	return &GroupPostgres{db: db}
}

func (r *GroupPostgres) Create(checkerId int, group classosbackend.Group) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createListQuery := fmt.Sprintf("INSERT INTO %s (name) VALUES ($1) RETURNING id", groupsTable)
	row := tx.QueryRow(createListQuery, group.Name)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	// здесь будет логика запроса в сервер AD через кмд или терминал
	//
	//снизу чтобы знать как работать с роллбек и коммитом и екзек
	//
	// createUsersListQuery := fmt.Sprintf("sadsadwa")
	// _, err = tx.Exec(createUsersListQuery, checkerId, id)
	// if err != nil {
	// 	tx.Rollback()
	// 	return 0, err
	// }

	return id, tx.Commit()
}

func (r *GroupPostgres) GetAll(checkerId int) ([]classosbackend.Group, error) {
	var groups []classosbackend.Group
	query := fmt.Sprintf("SELECT id, name FROM %s", groupsTable)

	err := r.db.Select(&groups, query)

	return groups, err
}

func (r *GroupPostgres) GetById(checkerId, groupId int) (classosbackend.Group, error) {
	var group classosbackend.Group

	query := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", groupsTable)

	err := r.db.Get(&group, query, groupId)

	return group, err
}

func (r *GroupPostgres) Delete(checkerId, groupId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", groupsTable)
	_, err := r.db.Exec(query, groupId)

	return err
}

func (r *GroupPostgres) Update(checkerId, groupId int, input classosbackend.UpdateGroupInput) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1
	
	if input.Name != nil {
		setValues = append(setValues, fmt.Sprintf("name=$%d", argId))
		args = append(args, *input.Name)
		argId++
	}

	setQuery := strings.Join(setValues, ", ")
	
	query := fmt.Sprintf("UPDATE %s Set %s Where id = $%d", groupsTable, setQuery, argId)
	args = append(args, groupId)

	_, err := r.db.Exec(query, args...)
	return err
}