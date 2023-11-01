# Tsing
Tsing是一个Go语言的HTTP API框架，具有以下功能特性：
- 高性能，零内存分配
- 微核心，仅包含路由和会话管理两大模块
- 轻量，无第三方包依赖，兼容`net/http`标准包
- 可自动处理路由处理器中的`Panic`错误，防止进程退出
- 使用回调函数代替传统的内置`Logger`机掉，异常处理更灵活
- 支持后置回调处理器`AfterHandler`（仅在路由命中时有效）

`Tsing`是汉字【青】以及同音字做为名词时的英文，例如：清华大学(Tsinghua University)、青岛(Tsing Tao)。

已经在多个项目中稳定运行。

## 执行流程
1. 根据`URI`查找路由
2. 执行`Use()`方法注册的中间件
3. 执行`GET()`等方法注册的路由处理器
4. 执行`AfterHandler`后置处理器，无视之前处理器的执行结果

- 如果任意环节的处理器出现了以下情况，会中止执行后面的处理器
    - 处理器返回了`error`
    - 处理器中执行了`Context.Break()`
    - 处理器执行时触发了`panic`
- 如果路由未命中，只会执行`ErrorHandler`错误回调处理器，不会触发中间件和后置处理器

## 安装
要求：Go 1.18+
```
github.com/dxvgef/tsing/v2
```

## 示例
请参考[/example_test.go](https://github.com/dxvgef/tsing/blob/master/example_test.go)文件

## 基准测试

测试代码：[github.com/dxvgef/tsing-benchmark](https://github.com/dxvgef/tsing-benchmark)

```
Benchmark_TsingV2-8                        50865             23725 ns/op               0 B/op          0 allocs/op
Benchmark_TsingV2_Recover-8                48708             24582 ns/op               0 B/op          0 allocs/op
Benchmark_TsingV1-8                        48664             24875 ns/op               0 B/op          0 allocs/op
Benchmark_TsingV1_Recover-8                45986             26267 ns/op               0 B/op          0 allocs/op
Benchmark_Gin-8                            47978             24542 ns/op               0 B/op          0 allocs/op
Benchmark_Gin_Recover-8                    43753             27390 ns/op               0 B/op          0 allocs/op
Benchmark_Httprouter-8                     46738             25555 ns/op           13792 B/op        167 allocs/op
Benchmark_Httprouter_Recover-8             44786             26703 ns/op           13792 B/op        167 allocs/op
Benchmark_Echo-8                           38401             31216 ns/op               0 B/op          0 allocs/op
Benchmark_Echo_Recover-8                   28674             41750 ns/op            9748 B/op        203 allocs/op
Benchmark_HTTPTreemux-8                    15448             77755 ns/op           65857 B/op        671 allocs/op
```

## 相关资源

- [dxvgef/tsing-demo](https://github.com/dxvgef/tsing-demo) `Tsing`整合常见功能的示例项目，可以做为新项目初始化使用
- [dxvgef/filter](https://github.com/dxvgef/filter) 参数值过滤包，由数据输入、格式化、校验、输出几个部份组成
- [Tsing Gateway](https://github.com/dxvgef/tsing-gateway) 跨平台、去中心化集群、动态配置的API网关
- [Tsing Center](https://github.com/dxvgef/tsing-center) 跨平台、去中心化集群、动态配置的服务中心

## 用户案例

如果你在使用本项目，请通过[Issues](https://github.com/dxvgef/tsing/issues)告知我们项目的简介

## 帮助反馈

本项目已在多个项目的生产环境中稳定运行。如有问题可在[Issues](https://github.com/dxvgef/tsing/issues)里提出。

诚邀更多的开发者参与到本项目维护中，帮助这个开源项目更好的发展。
