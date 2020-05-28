package main

import (
	"net/http"

	"github.com/dxvgef/tsing"
)

func main() {
	engine := tsing.New(&tsing.Config{})

	engine.GET("/", func(context *tsing.Context) error {
		context.ResponseWriter.Write([]byte("hello world"))
		return nil
	})
	http.ListenAndServe(":5656", engine)
}
