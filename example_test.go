package tsing

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 错误回调处理器
func errorHandler(ctx *Context) {
	log.Println("错误:", ctx.Status, ctx.Error)
	// 自动响应OPTIONS请求，用于解决CORS问题
	if ctx.Status == http.StatusMethodNotAllowed && ctx.Request.Method == "OPTIONS" {
		ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", ctx.Request.Header.Get("Origin"))
		ctx.ResponseWriter.Header().Set("Vary", "Origin")
		ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "*")
		ctx.ResponseWriter.Header().Set("Access-Control-Expose-Headers", "*")
		ctx.ResponseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.ResponseWriter.Header().Set("Access-Control-Max-Age", "2592000")
		ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
		return
	}

	// 常规处理错误
	ctx.ResponseWriter.WriteHeader(ctx.Status)
	if ctx.Error != nil {
		_, _ = ctx.ResponseWriter.Write([]byte(ctx.Error.Error()))
	}
}

// 测试回应
func TestEcho(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		t.Log("Hello Tsing")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试处理器执行顺序
func TestHandlers(t *testing.T) {
	app := New(Config{
		AfterHandlerFirstInFirstOut: true, // 后置处理器先注册先执行，否则先注册后执行
	})
	app.Before(func(ctx *Context) error {
		t.Log("1 执行了全局前置处理器")
		return nil
	})
	app.After(func(ctx *Context) error {
		t.Log("6 执行了全局后置处理器")
		return nil
	})
	group := app.Group("/group", func(ctx *Context) error {
		t.Log("2 执行了路由组/group 处理器")
		return nil
	})
	group.Before(func(ctx *Context) error {
		t.Log("3 执行了路由组/group 前置处理器")
		return nil
	})
	group.After(func(ctx *Context) error {
		t.Log("5 执行了路由组/group 后置处理器")
		return nil
	})
	group.GET("/object", func(ctx *Context) error {
		t.Log("4 执行了路由/group/object 处理器")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/group/object", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 PathValue
func TestPathValue(t *testing.T) {
	app := New()
	app.GET("/:path/:file", func(ctx *Context) error {
		t.Log("path=", ctx.PathValue("path"))
		t.Log("file=", ctx.PathValue("file"))
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/haha/123", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试Context传值
func TestContextValue(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		// 在ctx中写入参数
		ctx.SetValue("hello", "tsing")
		return nil
	}, func(ctx *Context) error {
		t.Log("hello=", ctx.GetValue("hello"))
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试中止处理器链
func TestBreak(t *testing.T) {
	app := New()
	group := app.Group("/group")
	group.GET("/object", func(ctx *Context) error {
		t.Log("ok")
		ctx.Break()
		return nil
	}, func(ctx *Context) error {
		t.Error("中止失败")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/group/object", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试QueryValue
func TestQueryParams(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		t.Log("id=", ctx.QueryValue("id"))
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/?id=123", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试FormValue
func TestFormValue(t *testing.T) {
	app := New()
	app.POST("/", func(ctx *Context) error {
		t.Log("test=", ctx.FormValue("test"))
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "POST", "/", strings.NewReader("test=ok"))
	if err != nil {
		t.Error(err)
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试404错误
func TestNotFoundError(t *testing.T) {
	app := New(Config{
		ErrorHandler: errorHandler,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/404", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试405事件
func TestMethodNotAllowedError(t *testing.T) {
	app := New(Config{
		HandleMethodNotAllowed: true,
		ErrorHandler:           errorHandler,
	})
	app.POST("/", func(ctx *Context) error {
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试panic事件
func TestPanicError(t *testing.T) {
	app := New(Config{
		Recovery:     true,
		ErrorHandler: errorHandler,
	})
	app.GET("/", func(ctx *Context) error {
		panic("panic消息")
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
