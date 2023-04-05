package main

import (
	"github.com/Abhra303/quickmap/pkg/datastore"
	"github.com/gin-gonic/gin"
)

var Store datastore.DataStore

func main() {
	Store = datastore.NewDataStore()
	// start server
	router := gin.Default()

	router.GET("api/v1/exec-command", v1ExecCommandHandler)

	router.Run(":5050")
}
