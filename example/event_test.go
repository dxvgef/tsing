package example

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dxvgef/tsing"
)

func TestEvent(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	app := tsing.New(&tsing.Config{
		RedirectTrailingSlash: false,
		HandleOPTIONS:         false,
		Recover:               true,
		ErrorEvent:            true,
		Trigger:               true,
		Trace:                 false,
		ShortPath:             true,
		EventHandler: func(event *tsing.Event) {
			log.Println(event.Message)
			if event.Trigger != nil {
				log.Println(event.Trigger.File)
				log.Println(event.Trigger.Line)
				log.Println(event.Trigger.Func)
			}
			log.Println(event.Trace)
		},
	})
	app.Router.GET("/return", func(ctx *tsing.Context) error {
		return ctx.Error(errors.New("return消息"))
	})

	app.Router.GET("/panic", func(ctx *tsing.Context) error {
		panic("panic消息")
		// nolint:govet
		return nil
	})

	// nolint:staticcheck
	r, err := http.NewRequest("GET", "/return", nil)
	if err != nil {
		t.Error(err.Error())
		// return
	}
	r, err = http.NewRequest("GET", "/panic", nil)
	if err != nil {
		t.Error(err.Error())
		// return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
