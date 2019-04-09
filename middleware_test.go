package tsing

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	app := New()
	app.Event.Handler = func(event Event) {
		log.Println(event.Message)
	}
	router := app.Router.GROUP("", func(ctx Context) (Context, error) {
		log.Println("执行了路由组中间件AAA")
		return ctx.Continue()
	}, func(ctx Context) (Context, error) {
		log.Println("执行了路由组中间件BBB")
		return ctx.Continue()
	})
	router.GET("/", func(ctx Context) error {
		log.Println("执行了最终的路由函数")
		return nil
	}, func(ctx Context) (Context, error) {
		log.Println("执行了路由中间件CCC")
		return ctx.Continue()
	}, func(ctx Context) (Context, error) {
		log.Println("执行了路由中间件DDD")
		return ctx.Continue()
	})

	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery
	r.Method = "GET"
	r.RequestURI = "/"
	u.Path = "/"
	u.RawQuery = rq
	app.ServeHTTP(httptest.NewRecorder(), r)
}
