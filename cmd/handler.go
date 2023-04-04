package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Command string `json:"command"`
}

type response struct {
	Code  int    `json:"code"`
	Value string `json:"value,omitempty"`
	Err   string `json:"error,omitempty"`
}

func v1ExecCommandHandler(c *gin.Context) {
	var req request
	var err error

	if err = c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &response{Code: http.StatusBadRequest, Err: err.Error()})
	}
}
