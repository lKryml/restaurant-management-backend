package main

import (
	"fmt"
	"os"
	"restaurant-management-backend/internal/logger"
	"restaurant-management-backend/internal/server"
)

func main() {
	logger.InitLogger()

	server := server.NewServer()

	logger.Log.Info(fmt.Sprintf("Server started on port:%s", os.Getenv("PORT")))

	err := server.ListenAndServe()
	if err != nil {
		logger.Log.Fatal("Failed to start server: ", err)
		//panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
