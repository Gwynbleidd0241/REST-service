package main

import (
	_ "effective_mobile/api"
	"effective_mobile/internal/config"
	"effective_mobile/internal/db"
	"effective_mobile/internal/handler"
	"effective_mobile/internal/logger"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
)

// @title Subscriptions API
// @version 1.1
// @description REST API для управления подписками
// @host 195.133.20.34:8081
// @BasePath /
func main() {
	config.LoadConfig()
	logger.Init()
	logger.Log.Info("Логгер инициализирован")

	dbConn, err := db.NewPostgres(
		config.Cfg.DBHost,
		config.Cfg.DBPort,
		config.Cfg.DBUser,
		config.Cfg.DBPassword,
		config.Cfg.DBName,
	)
	if err != nil {
		logger.Log.Fatalf("Ошибка подключения к базе: %v", err)
	}
	defer dbConn.Conn.Close()

	logger.Log.Info("Подключение к базе данных установлено")

	r := mux.NewRouter()
	h := handler.New(dbConn.Conn)
	r.HandleFunc("/subscriptions", h.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions", h.GetAllSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions/total", h.GetTotalPrice).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.GetSubscriptionByID).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.UpdateSubscription).Methods("PUT")
	r.HandleFunc("/subscriptions/{id}", h.DeleteSubscription).Methods("DELETE")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	logger.Log.Infof("Сервер запущен на порту %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
