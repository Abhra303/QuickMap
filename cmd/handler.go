package main

import (
	"net/http"

	"github.com/Abhra303/quickmap/pkg/parse"
	"github.com/efficientgo/core/errors"
	"github.com/gin-gonic/gin"
)

type request struct {
	Command string `json:"command"`
}

type response struct {
	Code  int         `json:"code"`
	Value interface{} `json:"value,omitempty"`
	Err   string      `json:"error,omitempty"`
}

func badReqResponse(c *gin.Context, err error) {
	c.Error(err)
	c.AbortWithStatusJSON(http.StatusBadRequest, &response{Code: http.StatusBadRequest, Err: err.Error()})
}

func v1ExecCommandHandler(c *gin.Context) {
	var req request
	var err error

	if err = c.BindJSON(&req); err != nil {
		badReqResponse(c, err)
		return
	}

	expr, err := parse.ParseCommand(req.Command)
	if err != nil {
		badReqResponse(c, err)
		return
	}
	switch e := expr.(type) {
	case *parse.SetExpr:
		created, err := Store.Set(e.Key, e.Value, e.Expiry, e.Condition)
		if err != nil {
			badReqResponse(c, err)
			return
		}
		if created {
			c.JSON(http.StatusCreated, &response{Code: http.StatusCreated})
		} else {
			c.JSON(http.StatusOK, &response{Code: http.StatusOK})
		}
	case *parse.GetExpr:
		value := Store.Get(e.Key)
		c.JSON(http.StatusOK, &response{Code: http.StatusOK, Value: value})
	case *parse.QPushExpr:
		err = Store.QPush(e.Key, e.Values)
		if err != nil {
			badReqResponse(c, err)
			return
		}
		c.JSON(http.StatusCreated, &response{Code: http.StatusCreated})
	case *parse.QPopExpr:
		value, err := Store.QPop(e.Key)
		if err != nil {
			badReqResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, &response{Code: http.StatusOK, Value: value})
	case *parse.BQPopExpr:
		value, err := Store.BQPop(e.Key, e.Timeout)
		if err != nil {
			badReqResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, &response{Code: http.StatusOK, Value: value})
	default:
		err = errors.New("unknown expression type")
		badReqResponse(c, err)
		return
	}
}
