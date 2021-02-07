# Tsing
Tsing是一个Go语言的HTTP API框架，具有以下优秀的特性：
- 高性能，零内存分配
- 微核心，方便二次开发
- 轻量，无第三方包依赖，兼容net/http标准包
- 统一异常处理，减少代码量，使开发者专注于业务逻辑

Tsing诞生比较晚也没有刻意的宣传推广，但由于框架核心非常精简，可靠性并不会低于其它热门框架，并且基准测试结果证明它的性能要强于所有参与测试的主流框架，已经在多个未公开的项目中稳定运行。

Tsing 是汉字【青】以及同音字做为名词时的英文，例如：清华大学(Tsinghua University)、青岛(Tsing Tao)。

> github.com/dxvgef/tsing

## 手册：

* [基本示例](https://github.com/dxvgef/tsing/wiki/%E5%9F%BA%E6%9C%AC%E7%A4%BA%E4%BE%8B)
* [异常处理](https://github.com/dxvgef/tsing/wiki/%E5%BC%82%E5%B8%B8%E5%A4%84%E7%90%86)
* [路由及路由处理器](https://github.com/dxvgef/tsing/wiki/%E8%B7%AF%E7%94%B1%E5%8F%8A%E8%B7%AF%E7%94%B1%E5%A4%84%E7%90%86%E5%99%A8)
* [会话Context](https://github.com/dxvgef/tsing/wiki/%E4%BC%9A%E8%AF%9DContext)
* [参数验证及类型转换](https://github.com/dxvgef/tsing/wiki/%E5%8F%82%E6%95%B0%E9%AA%8C%E8%AF%81%E5%8F%8A%E7%B1%BB%E5%9E%8B%E8%BD%AC%E6%8D%A2)
* [HTML模板渲染](https://github.com/dxvgef/tsing/wiki/HTML%E6%A8%A1%E6%9D%BF%E6%B8%B2%E6%9F%93)
* [CORS跨域资源共享控制](https://github.com/dxvgef/tsing/wiki/CORS%E8%B7%A8%E5%9F%9F%E8%B5%84%E6%BA%90%E5%85%B1%E4%BA%AB%E6%8E%A7%E5%88%B6)
* [Session](https://github.com/dxvgef/tsing/wiki/Session)
* [JWT(JSON Web Token)](https://github.com/dxvgef/tsing/wiki/JSON-Web-Token)
* [事件记录](https://github.com/dxvgef/tsing/wiki/%E4%BA%8B%E4%BB%B6%E8%AE%B0%E5%BD%95)
* [优雅关闭(Graceful Shutdown)](https://github.com/dxvgef/tsing/wiki/%E4%BC%98%E9%9B%85%E5%85%B3%E9%97%AD(Graceful-Shutdown))

更多示例代码请参考[/example_test.go](https://github.com/dxvgef/tsing/blob/master/example_test.go)文件

## 基准测试

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
[dxvgef/tsing-benchmark](https://github.com/dxvgef/tsing-benchmark)是`tsing`,`httprouter`,`echo`,`gin`,`chi`等框架的基准测试代码，不定期更新各框架的版本或加入新的框架进行测试


## 相关资源

- [dxvgef/tsing-demo](https://github.com/dxvgef/tsing-demo) `Tsing`整合常见功能的示例项目，可以做为新项目初始化使用
- [Tsing Gateway](https://github.com/dxvgef/tsing-gateway) 开源、跨平台、去中心化集群、动态配置的API网关
- [Tsing Center](https://github.com/dxvgef/tsing-center) 开源、跨平台、去中心化集群、动态配置的服务中心

## 用户及案例

如果你在使用本项目，请通过[Issues](https://github.com/dxvgef/tsing/issues)告知我们项目的简介

## 帮助/说明

本项目已在多个项目的生产环境中稳定运行。如有问题可在[Issues](https://github.com/dxvgef/tsing/issues)里提出。

诚邀更多的开发者参与到本项目维护中，帮助这个开源项目更好的发展。
