package example

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dxvgef/tsing"
)

func TestMiddleware(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	app := tsing.New(&tsing.Config{
		RedirectTrailingSlash: false,
		HandleOPTIONS:         false,
		Recover:               true,
		ErrorEvent:            true,
		NotFoundEvent:         true,
		MethodNotAllowedEvent: true,
		Trigger:               true,
		Trace:                 true,
		ShortPath:             true,
		EventHandler: func(event *tsing.Event) {
			log.Println(event.Message)
			log.Println(event.Trigger)
			log.Println(event.Trace)
		},
	})
	router := app.Router.GROUP("", func(ctx *tsing.Context) error {
		ctx.SetValue("test", "这是路由组中间件传递下来的消息")
		t.Log("执行了路由组中间件AAA")
		ctx.Next()
		return nil
	}, func(ctx *tsing.Context) error {
		t.Log("执行了路由组中间件BBB")
		ctx.Next()
		return nil
	})
	router.POST("/", func(ctx *tsing.Context) error {
		log.Println("执行了路由处理器")
		log.Println(ctx.GetValue("test"))
		return nil
	}, func(ctx *tsing.Context) error {
		log.Println("执行了路由中间件CCC")
		ctx.Next()
		return nil
	}, func(ctx *tsing.Context) error {
		log.Println("执行了路由中间件DDD")
		ctx.Next()
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
