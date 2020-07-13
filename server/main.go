/*
 * Simple blogging APIs
 *
 * This is a simple blogging API
 *
 */

package main

import (
	serve "github.com/gouthams/blogApp/server/restimpl"
	"github.com/gouthams/blogApp/server/utils"
)

const port = ":8080"

func main() {
	//Initialize logging framework
	utils.InitializeLogging()

	//Initialize DB
	utils.ConnectToDatabase()
	logEntry := utils.Log()
	router := serve.NewRouter()

	err := router.Run(port)
	if err != nil {
		logEntry.Fatalf("Unable to start the server on port:%s", port)
	}
	logEntry.Info("Server started on port:%s", port)
}
