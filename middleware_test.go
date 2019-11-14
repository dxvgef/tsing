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
	app.Config.EventHandler = func(event *Event) {
		log.Println(event.Message)
	}
	router := app.Router.GROUP("", func(ctx *Context) error {
		ctx.SetValue("test", "这是路由组中间件传递下来的数据")
		t.Log("执行了路由组中间件AAA")
		ctx.Next()
		return nil
	}, func(ctx *Context) error {
		t.Log("执行了路由组中间件BBB")
		// ctx.Next()
		return nil
	})
	router.POST("/", func(ctx *Context) error {
		log.Println("执行了[" + ctx.Request.Method + "]" + ctx.Request.URL.Path + "方法")
		log.Println(ctx.GetValue("test"))
		return nil
	}, func(ctx *Context) error {
		log.Println("执行了路由中间件CCC")
		return nil
	}, func(ctx *Context) error {
		log.Println("执行了路由中间件DDD")
		return nil
	})

	v := url.Values{}
	v.Add("email", "dxvgef@outlook.com")
	r, err := http.NewRequest("POST", "/", strings.NewReader(v.Encode()))
	if err != nil {
		log.Println(err.Error())
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}
