# Tsing
Tsing是一个Go语言的Web框架，具有以下优秀的特性：
- 高性能，零内存分配
- 微核心，方便二次开发
- 轻量，无第三方包依赖，兼容net/http标准包
- 统一异常处理，减少代码量，使开发者专注于业务逻辑


### 安装：
> go get -u github.com/dxvgef/tsing

### 手册：
* [基本示例](https://github.com/dxvgef/tsing/wiki/%E5%9F%BA%E6%9C%AC%E7%A4%BA%E4%BE%8B)
* [异常处理](https://github.com/dxvgef/tsing/wiki/%E5%BC%82%E5%B8%B8%E5%A4%84%E7%90%86)
* [路由及路由处理器](https://github.com/dxvgef/tsing/wiki/%E8%B7%AF%E7%94%B1%E5%8F%8A%E8%B7%AF%E7%94%B1%E5%A4%84%E7%90%86%E5%99%A8)
* [会话Context](https://github.com/dxvgef/tsing/wiki/%E4%BC%9A%E8%AF%9DContext)
* [参数验证及类型转换](https://github.com/dxvgef/tsing/wiki/%E5%8F%82%E6%95%B0%E9%AA%8C%E8%AF%81%E5%8F%8A%E7%B1%BB%E5%9E%8B%E8%BD%AC%E6%8D%A2)
* [数据输出](https://github.com/dxvgef/tsing/wiki/%E6%95%B0%E6%8D%AE%E8%BE%93%E5%87%BA)
* [HTML模板渲染](https://github.com/dxvgef/tsing/wiki/HTML%E6%A8%A1%E6%9D%BF%E6%B8%B2%E6%9F%93)
* [CORS跨域资源共享控制](https://github.com/dxvgef/tsing/wiki/CORS%E8%B7%A8%E5%9F%9F%E8%B5%84%E6%BA%90%E5%85%B1%E4%BA%AB%E6%8E%A7%E5%88%B6)
* [Session](https://github.com/dxvgef/tsing/wiki/Session)
* [JWT(JSON Web Token)](https://github.com/dxvgef/tsing/wiki/JSON-Web-Token)
* [事件Logger](https://github.com/dxvgef/tsing/wiki/%E4%BA%8B%E4%BB%B6Logger)
* [优雅关闭(Graceful Shutdown)](https://github.com/dxvgef/tsing/wiki/%E4%BC%98%E9%9B%85%E5%85%B3%E9%97%AD(Graceful-Shutdown))

更多示例代码请参考[/example_test.go](https://github.com/dxvgef/tsing/blob/master/example_test.go)文件

### 完整的项目示例
[dxvgef/tsing-demo](https://github.com/dxvgef/tsing-demo) 是一个基于`Tsing`整合常见功能的完整示例，可以做为项目初始化使用

### 基准测试
`tsing`与`httprouter`、`echo`、`gin`等框架的基准测试对比
<br>[dxvgef/tsing-benchmark](https://github.com/dxvgef/tsing-benchmark)

测试结果：
```
Benchmark_Tsing_V1-4                       42688             26372 ns/op               0 B/op          0 allocs/op
Benchmark_Tsing_V1_Recover-4               41553             27571 ns/op               0 B/op          0 allocs/op
Benchmark_Httprouter-4                     33806             32360 ns/op           13792 B/op        167 allocs/op
Benchmark_Httprouter_Recover-4             35547             33129 ns/op           13792 B/op        167 allocs/op
Benchmark_Gin-4                            33469             34294 ns/op            6497 B/op        203 allocs/op
Benchmark_Gin_Recover-4                    31071             37423 ns/op            6497 B/op        203 allocs/op
Benchmark_Echo-4                           31489             36706 ns/op               0 B/op          0 allocs/op
Benchmark_Echo_Recover-4                   21991             53318 ns/op            9745 B/op        203 allocs/op
```