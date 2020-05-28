package main

import (
	"net/http"

	"github.com/dxvgef/tsing"
	"github.com/dxvgef/tsing/middleware"
)

func main() {
	engine := tsing.New(&tsing.Config{})

	basicAuth := middleware.BasicAuth("Restricted", map[string]string{"admin": "admin"})

	engine.GET("/", basicAuth, func(context *tsing.Context) error {
		context.ResponseWriter.Write([]byte("hello world"))
		return nil
	})

	http.ListenAndServe(":5657", engine)
}
