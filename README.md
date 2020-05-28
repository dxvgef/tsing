# Tsing
Tsing is a lightweight, high performance and no dependency router for building Go HTTP services


### [中文手册](README_ZH.md)
### Feature
- **lightweight** - no dependency third's package，compatible with net/http package
- **high performance** - memory zero allocation
- **simple** - convenient for secondary development
- **middleware** - support middleware code


### Install
> go get -u github.com/dxvgef/tsing

### Manual
* [Basic Example](https://github.com/dxvgef/tsing/wiki/%E5%9F%BA%E6%9C%AC%E7%A4%BA%E4%BE%8B)
* [Exception Handing](https://github.com/dxvgef/tsing/wiki/%E5%BC%82%E5%B8%B8%E5%A4%84%E7%90%86)
* [Route](https://github.com/dxvgef/tsing/wiki/%E8%B7%AF%E7%94%B1%E5%8F%8A%E8%B7%AF%E7%94%B1%E5%A4%84%E7%90%86%E5%99%A8)
* [Context](https://github.com/dxvgef/tsing/wiki/%E4%BC%9A%E8%AF%9DContext)
* [Param Handing](https://github.com/dxvgef/tsing/wiki/%E5%8F%82%E6%95%B0%E9%AA%8C%E8%AF%81%E5%8F%8A%E7%B1%BB%E5%9E%8B%E8%BD%AC%E6%8D%A2)
* [Data Output](https://github.com/dxvgef/tsing/wiki/%E6%95%B0%E6%8D%AE%E8%BE%93%E5%87%BA)
* [HTML Render](https://github.com/dxvgef/tsing/wiki/HTML%E6%A8%A1%E6%9D%BF%E6%B8%B2%E6%9F%93)
* [CORS](https://github.com/dxvgef/tsing/wiki/CORS%E8%B7%A8%E5%9F%9F%E8%B5%84%E6%BA%90%E5%85%B1%E4%BA%AB%E6%8E%A7%E5%88%B6)
* [Session](https://github.com/dxvgef/tsing/wiki/Session)
* [JWT(JSON Web Token)](https://github.com/dxvgef/tsing/wiki/JSON-Web-Token)
* [Event & Logger](https://github.com/dxvgef/tsing/wiki/%E4%BA%8B%E4%BB%B6Logger)
* [Graceful Shutdown](https://github.com/dxvgef/tsing/wiki/%E4%BC%98%E9%9B%85%E5%85%B3%E9%97%AD(Graceful-Shutdown))

more example,please see [example_test.go](https://github.com/dxvgef/tsing/blob/master/example_test.go)


### Tsing examples

* [hello_world](https://github.com/dxvgefi/tsing/blob/master/_examples/hello_world/main.go) - Hello World!
* [graceful](https://github.com/dxvgef/tsing/blob/master/_examples/graceful/main.go) - Graceful context signaling and server shutdown
* [basic_auth](https://github.com/dxvgef/tsing/blob/master/_examples/basic_auth/main.go) - Handler With simple Basic Auth



### Project Example
[dxvgef/tsing-demo](https://github.com/dxvgef/tsing-demo) It is a complete example of integrating common functions based on `Tsing`, which can be used for project initialization

### Benchmark
Benchmark comparison results of `Tsing` and `httprouter`,`echo` and `gin`.
<br>[dxvgef/tsing-benchmark](https://github.com/dxvgef/tsing-benchmark)

Result:
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