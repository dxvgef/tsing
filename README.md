# Tsing
Tsing是一个Go语言的HTTP API框架，具有以下功能特性：
- 高性能，零内存分配
- 微核心，仅包含路由和会话管理两大模块
- 轻量，无第三方包依赖，兼容`net/http`标准包
- 自动恢复路由处理器中的`Panic`错误，防止进程退出
- 错误回调函数机制代替内置`Logger`，异常处理更灵活
- 路由处理器支持前/后置钩子（仅在路由命中时有效）
- 后置钩子处理器支持**先进先出**和**先进后出**两种执行顺序

`Tsing`是汉字【青】以及同音字做为名词时的英文，例如：清华大学(Tsinghua University)、青岛(Tsing Tao)。

`Tsing`诞生比较晚也没有刻意的宣传推广，但由于框架核心非常精简，可靠性并不会低于其它热门框架，并且基准测试结果证明它的性能要强于所有参与测试的主流框架，已经在多个项目中稳定运行。

## 执行流程
1. 根据`URI`查找路由
2. 执行`Beforc()`方法注册的前置处理器
3. 执行`GET()`等执行方式注册的路由处理器
4. 执行`After()`方法注册的后置处理器

- 如果任意环节的处理器出现了以下情况，会中止执行后面的处理器
    - 处理器返回了`error`
    - 处理器中执行了`Context.Break()`
    - 处理器执行时触发了`panic`
- 如果路由未匹配，只执行`ErrorHandler`错误回调处理器，不会触发前置和后置处理器。

## 安装
要求：Go 1.18+
```
github.com/dxvgef/tsing/v2
```

## 示例
请参考[/example_test.go](https://github.com/dxvgef/tsing/blob/master/example_test.go)文件

## 基准测试

[github.com/dxvgef/tsing-benchmark](https://github.com/dxvgef/tsing-benchmark)

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
