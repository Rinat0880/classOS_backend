package repository

type Authorization interface {
}

type Group interface {
}

type User interface {
}

type Repository struct {
	Authorization
	Group
	User
}

func NewRepository() *Repository {
	return &Repository{}
}
