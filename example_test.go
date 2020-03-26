package tsing

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// 测试回应
func TestEcho(t *testing.T) {
	app := Default()
	app.GET("/", func(ctx *Context) {
		ctx.ResponseWriter.WriteHeader(200)
		ctx.ResponseWriter.Write([]byte("Hello !"))
		t.Log("Hello !")
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 PathParams
func TestURLParams(t *testing.T) {
	app := Default()
	app.GET("/:test/ok", func(ctx *Context) {
		t.Log(ctx.URLParams.Value("test"))
	})
	r, err := http.NewRequest("GET", "/haha/ok", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试Context传值
func TestContext(t *testing.T) {
	app := Default()
	app.GET("/context", func(ctx *Context) {
		// 在ctx中写入参数
		ctx.SetValue("test", "hehe")
		t.Log(1, ctx.Request.URL.Path, "写值")
	}, func(ctx *Context) {
		// 从ctx中读取参数
		t.Log(2, ctx.Request.URL.Path, "取值：", ctx.GetValue("test"))
	})
	r, err := http.NewRequest("GET", "/context", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试路由组
func TestGroup(t *testing.T) {
	app := Default()
	group := app.Group("/group", func(ctx *Context) {
		ctx.SetValue("test", "haha")
		t.Log(1, ctx.Request.URL.Path, "写值")
	})
	group.GET("/object", func(ctx *Context) {
		t.Log(2, ctx.Request.URL.Path, "取值：", ctx.GetValue("test"))
	}, func(ctx *Context) {
		t.Log(3, ctx.Request.URL.Path, "取值：", ctx.GetValue("test"))
	})
	r, err := http.NewRequest("GET", "/group/object", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试中止
func TestAbort(t *testing.T) {
	app := Default()
	group := app.Group("/group")
	group.GET("/object", func(ctx *Context) {
		t.Log(1, ctx.Request.URL.Path)
		ctx.Abort()
		t.Log(2, ctx.IsAborted())
	}, func(ctx *Context) {
		t.Log(3, ctx.Request.URL.Path)
	})
	r, err := http.NewRequest("GET", "/group/object", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试添加处理器
func TestAppend(t *testing.T) {
	app := Default()
	group := app.Group("/group")
	group.Append(func(ctx *Context) {
		t.Log(1, "append handler 1")
	}, func(ctx *Context) {
		t.Log(2, "append handler 2")
	})
	group.GET("/object", func(ctx *Context) {
		t.Log(3, ctx.Request.URL.Path)
	})
	r, err := http.NewRequest("GET", "/group/object", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试QueryValues
func TestQueryValues(t *testing.T) {
	app := Default()
	app.GET("/object", func(ctx *Context) {
		t.Log(ctx.QueryValues())
	})
	r, err := http.NewRequest("GET", "/object?a=1&b=2", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试PostValues
func TestPostValues(t *testing.T) {
	app := Default()
	app.POST("/object", func(ctx *Context) {
		t.Log(ctx.PostValues())
	})

	v := url.Values{}
	v.Add("a", "1")
	v.Add("b", "2")
	r, err := http.NewRequest("POST", "/object", strings.NewReader(v.Encode()))
	if err != nil {
		t.Error(err.Error())
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试FormValues
func TestFormValues(t *testing.T) {
	app := Default()
	app.POST("/object", func(ctx *Context) {
		t.Log(ctx.FormValues())
	})

	v := url.Values{}
	v.Add("c", "3")
	v.Add("d", "4")
	r, err := http.NewRequest("POST", "/object?a=1&b=2", strings.NewReader(v.Encode()))
	if err != nil {
		t.Error(err.Error())
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 事件处理器
func eventHandler(e *Event) {
	log.SetFlags(log.Lshortfile)
	log.Println(e.Status)
	log.Println(e.Message)
	log.Println(e.Source)
	for k := range e.Trace {
		log.Println("  ", e.Trace[k])
	}
}

// 测试panic事件
func TestPanicEvent(t *testing.T) {
	app := New(&Config{
		RootPath:              getRootPath(),
		UnescapePathValues:    true,
		MaxMultipartMemory:    2 << 20,
		EventHandler:          eventHandler,
		EventTrace:            true,
		EventTraceOnlyProject: true,
		EventTrigger:          true,
		Recover:               true,
		EventTraceShortPath:   true,
	})
	app.GET("/", func(ctx *Context) {
		panic("这是panic消息")
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试404事件
func TestNotFoundEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		EventHandler:       eventHandler,
	})
	r, err := http.NewRequest("GET", "/404", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试405事件
func TestMethodNotAllowedEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		EventHandler:       eventHandler,
	})
	app.POST("/", func(ctx *Context) {})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
