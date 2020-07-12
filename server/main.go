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
	"log"
)

const port = ":8080"

func main() {
	//Initialize logging framework
	utils.InitializeLogging()

	//Initialize DB
	utils.ConnectToDatabase()

	log.Printf("Server started")

	router := serve.NewRouter()

	log.Fatal(router.Run(port))
}
