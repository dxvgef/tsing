package tsing

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMiddleware(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	app := New()
	app.Event.Handler = func(event Event) {
		log.Println(event.Message)
	}
	router := app.Router.GROUP("", func(ctx Context) (Context, error) {
		ctx.SetContextValue("test", "这是路由组中间件传递下来的数据")
		log.Println("执行了路由组中间件AAA")
		return ctx.Continue()
	}, func(ctx Context) (Context, error) {
		log.Println("执行了路由组中间件BBB")
		return ctx.Continue()
	})
	router.POST("/", func(ctx Context) error {
		log.Println("执行了[" + ctx.Request.Method + "]" + ctx.Request.URL.Path + "方法")
		log.Println(ctx.ContextValue("test"))
		return nil
	}, func(ctx Context) (Context, error) {
		log.Println("执行了路由中间件CCC")
		return ctx.Continue()
	}, func(ctx Context) (Context, error) {
		log.Println("执行了路由中间件DDD")
		return ctx.Continue()
	})

	v := url.Values{}
	v.Add("email", "dxvgef@outlook.com")
	r, err := http.NewRequest("POST", "/", strings.NewReader(v.Encode()))
	if err != nil {
		log.Println(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
