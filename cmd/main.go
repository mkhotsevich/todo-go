package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"

	"github.com/spf13/viper"

	"github.com/mkhotsevich/todo-go/pkg/repository"
	"github.com/mkhotsevich/todo-go/pkg/service"

	"github.com/mkhotsevich/todo-go/pkg/handler"

	"github.com/mkhotsevich/todo-go"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := InitConfig(); err != nil {
		logrus.Fatalf("error initilizing cofing: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
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
		logrus.Fatalf("failed to initialize database: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(todo.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occurred while running http server: %s", err.Error())
		}
	}()

	logrus.Print("TodoApp started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("TodoApp shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Fatalf("error occurred on server shutting down: %s", err.Error())
	}
	if err := db.Close(); err != nil {
		logrus.Fatalf("error occurred on db connection close: %s", err.Error())
	}
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
