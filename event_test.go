package tsing

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvent(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	app := New()
	app.Config.EventTrigger = true
	app.Config.EventHandler = func(event *Event) {
		t.Log(event)
	}
	app.Router.GET("/", func(ctx *Context) error {
		return errors.New("错误消息")
	})

	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
