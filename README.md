# 第三代哈士齐营销平台 ( Gaea ) 后端
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/offcn-jl/gaea-back-end)](https://goreportcard.com/report/github.com/offcn-jl/gaea-back-end) [![持续交付](https://github.com/offcn-jl/gaea-back-end/workflows/CD/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACD) [![代码扫描](https://github.com/offcn-jl/gaea-back-end/workflows/CD%20CodeQL/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACD%20CodeQL) [![codecov](https://codecov.io/gh/offcn-jl/gaea-back-end/branch/main/graph/badge.svg)](https://codecov.io/gh/offcn-jl/gaea-back-end) [![持续集成](https://github.com/offcn-jl/gaea-back-end/workflows/CI/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACI) [![代码扫描](https://github.com/offcn-jl/gaea-back-end/workflows/CI%20CodeQL/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACI%20CodeQL) [![codecov](https://codecov.io/gh/offcn-jl/gaea-back-end/branch/new-feature/graph/badge.svg)](https://codecov.io/gh/offcn-jl/gaea-back-end/branch/new-feature) 

## 待办列表

### 基础架构
 - [ ] 完善单元测试 与 http 请求及响应有关的部分
 - [ ] 屏蔽 腾讯云API网关 的健康检查产生的无用的 HEAD 日志

### 内部服务接口 Services
 - [ ] 微信开放平台接口组
    - [ ] 授权事件接口
    - [ ] 授权方的消息与事件接口

### 管理平台接口 Manages
1. 重新梳理权限分级的逻辑
1. 重点摘出高资源消耗及高时间消耗的业务, 迁移到独立 pod 或 serverless 架构
    1. CRM 数据自动录入功能，摘出作为独立项目，由现在的长期运行，靠 sleep 实现定时轮询任务。修改为任务计划定期拉起 pod，这样可以在闲时节约资源。还可以测试迁移到 Serverless 的可能性。
    1. 同步粉丝列表功能, 从现在的收到请求就启动协程开始多线程处理 (无法预知同时进行的任务数量，可能导致某一时刻多个任务在多线程处理，导致数据库压力极大)，修改为定时查询任务列表，逐个进行多线程处理
    1. 个人后缀在新增时，检查目前生效的所有后缀中，是否存在重复的。 达到禁用时间后，立即失效。

### 外部活动接口 Events
1. 大部分服务依赖 Manages. 所以等待 Manages 开发完成后再进行规划.
1. 海报生成业务迁移到 serverless

## 单元测试

`参考 https://www.jianshu.com/p/e3b2b1194830`

1. 安装 GoConvey
    1. 在命令行执行  
        `go get github.com/smartystreets/goconvey`
    1. 运行时间较长，运行完后  
        1. 在$GOPATH/src目录下新增了github.com子目录，该子目录里包含了GoConvey框架的库代码
        1. 在$GOPATH/bin目录下新增了GoConvey框架的可执行程序goconvey
    
1. 运行 Web 界面 并 在 Web 界面进行自动化编译测试
    1. 在项目根目录执行  
       `$GOPATH/bin/goconvey`

1. 可以参考的内容
    1. 常用断言及用法 ( GoConvey断言err和bool的方法 ) https://blog.csdn.net/luomoshusheng/article/details/50226257?locationNum=13&fps=1

