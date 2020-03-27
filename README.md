# Tsing

Go语言的高性能Web框架，具有以下特点：
- 高性能，零内存分配
- 微核心，方便二次开发
- 轻量，无第三方包依赖
- 事件机制，自由处理异常
- 兼容net/http标准包，可配合大多数第三方包使用

#### 安装：
> go get -u github.com/dxvgef/tsing


#### 基本示例：

```go
app := New(&Config{})
app.GET("/", func(ctx *Context) error {
    ctx.ResponseWriter.WriteHeader(200)
    ctx.ResponseWriter.Write([]byte("Hello !"))
    return nil
})
if err := http.ListenAndServe(":8080", app); err != nil {
    log.Fatal(err.Error())
}
```

更多示例请参考`/example_test.go`文件

Wiki中的文档正在完善中...