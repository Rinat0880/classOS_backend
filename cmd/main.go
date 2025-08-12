package main

import (
	"log"

	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/handler"
	"github.com/rinat0880/classOS_backend/pkg/repository"
	"github.com/rinat0880/classOS_backend/pkg/service"
)

func main() {
	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(classosbackend.Server)
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running server: %s", err.Error())
	}
}
