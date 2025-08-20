package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	classosbackend "github.com/rinat0880/classOS_backend"
)

type GroupPostgres struct {
	db *sqlx.DB
}

func NewGroupPostgres(db *sqlx.DB) *GroupPostgres {
	return &GroupPostgres{db: db}
}

func (r *GroupPostgres) Create(userId int, group classosbackend.Group) (int, error) {
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
	// _, err = tx.Exec(createUsersListQuery, userId, id)
	// if err != nil {
	// 	tx.Rollback()
	// 	return 0, err
	// }

	return id, tx.Commit()
}

func (r *GroupPostgres) GetAll(userId int) ([]classosbackend.Group, error) {
	var groups []classosbackend.Group
	query := fmt.Sprintf("SELECT id, name FROM %s", groupsTable)

	err := r.db.Select(&groups, query)

	return groups, err
}

func (r *GroupPostgres) GetById(userId, groupId int) (classosbackend.Group, error) {
	var group classosbackend.Group

	query := fmt.Sprintf("SELECT id, name FROM %s WHERE id = $1", groupsTable)

	err := r.db.Get(&group, query, groupId)

	return group, err
}