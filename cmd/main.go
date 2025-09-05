package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	classosbackend "github.com/rinat0880/classOS_backend"
	"github.com/rinat0880/classOS_backend/pkg/handler"
	"github.com/rinat0880/classOS_backend/pkg/repository"
	"github.com/rinat0880/classOS_backend/pkg/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error in initializing config: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Printf("Warning: No .env file found: %v (using environment variables)", err)
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logrus.Fatalf("err in init db: %s", err.Error())
	}

	repos := repository.NewRepository(db)

	adService := service.NewADService()
	if err := adService.TestConnection(); err != nil {
		logrus.Printf("Warning: AD connection failed: %v", err)
		logrus.Printf("Application will work in DB-only mode")
	} else {
		logrus.Println("AD connection established successfully")
	}

	authService := service.NewAuthService(repos.Authorization)
	integratedUserService := service.NewIntegratedUserService(repos.User, repos.Group, authService, adService)

	services := &service.Service{
		Authorization: authService,
		Group: service.NewIntegratedGroupService(repos.Group, adService),
		User:          integratedUserService,
	}

	handlers := handler.NewHandler(services)

	srv := new(classosbackend.Server)
	if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		logrus.Fatalf("error occured while running server: %s", err.Error())
	}
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running server: %s", err.Error())
		}
	}()

	logrus.Print("classOS_backend started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("classOS_backend shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db conn closing: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
