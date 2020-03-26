package tsing

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// 测试 Hello
func TestHello(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) {
		t.Log("Hello !")
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 URLParams
func TestURLParams(t *testing.T) {
	app := New()
	app.GET("/:test/ok", func(ctx *Context) {
		t.Log(ctx.URLParams.Key("test"))
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
	app := New()
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
	app := New()
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
	app := New()
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
	app := New()
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
	app := New()
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
	app := New()
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
	app := New()
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
