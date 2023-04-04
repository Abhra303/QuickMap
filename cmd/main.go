package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// start server
	router := gin.Default()

	router.GET("api/v1/exec-command", v1ExecCommandHandler)

	router.Run(":5050")
}
