# 第三代哈士齐营销平台 ( Gaea ) 后端服务
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/offcn-jl/gaea-back-end)](https://goreportcard.com/report/github.com/offcn-jl/gaea-back-end) [![持续交付](https://github.com/offcn-jl/gaea-back-end/workflows/CD/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACD) [![代码扫描](https://github.com/offcn-jl/gaea-back-end/workflows/CD%20CodeQL/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACD%20CodeQL) [![codecov](https://codecov.io/gh/offcn-jl/gaea-back-end/branch/main/graph/badge.svg)](https://codecov.io/gh/offcn-jl/gaea-back-end) [![持续集成](https://github.com/offcn-jl/gaea-back-end/workflows/CI/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACI) [![代码扫描](https://github.com/offcn-jl/gaea-back-end/workflows/CI%20CodeQL/badge.svg)](https://github.com/offcn-jl/gaea-back-end/actions?query=workflow%3ACI%20CodeQL) [![codecov](https://codecov.io/gh/offcn-jl/gaea-back-end/branch/new-feature/graph/badge.svg)](https://codecov.io/gh/offcn-jl/gaea-back-end/branch/new-feature) 

## 待办列表
在原哈士齐营销平台的基础上，进行以下改进。

### 基础架构
- [x] 增加单元测试，在 CI/CD 流程中， 通过单元测试后再进行代码的编译及镜像的构建等操作
- [x] 增加 CodeQL 代码检查功能进行源码安全审查
- [x] 将分散在 TSF 、 Serverless Frame 架构中的模块重新整合; 将分散为三个的接入点 ( Services、Manage、Events ) 重新进行整合; 减少需要维护的项目数量, 增强项目的可维护性;  
    `目前的架构过于松散、混乱。尤其是部分接口和应用试用的 Serverless 架构经过验证后证明并不适合大面积应用于生产环境 ( 架构本身迭代速度过快，需要消耗大量经历适配他的最新迭代；稳定性不达标：故障率过高、故障后恢复时间过久。无SLA保障，出现故障后无赔偿 )`
- [x] 精简应用启动所需配置的环境变量数量，原项目中引入的环境变量 ( 用于配置 ) 过多，相互依赖复杂，配置项不清晰
- [x] 将配置项目迁移到数据库进行保存，实现每次配置更新可以在数据库中留有记录的效果
    - [ ] 应用繁忙时，有可能会拉起多个 Pod 。此时更新配置，会出现只有一个 Pod 更新到最新配置，其余 Pod 无法获取最新配置的问题
- [ ] 微信开放平台相关的 JSON 数据保存到 MongoDB 中
- [ ] 重新设计权限分级管理功能, 目前应用的权限分级管理功能不够成熟
    - [x] 角色、权限集成逻辑基本不变
    - [ ] 上级角色可以查看下级角色全部的资源 ( where gid in [ 0, 1, 2] )
    - [x] 取消前端右上角的角色切换操作，简化操作流程
- [ ] 摘出高资源消耗及高时间消耗的业务, 迁移到独立 pod 或 serverless 架构  
    `目前部分高资源消耗及高时间消耗的业务设计不合理，与 web 接口业务混合在一起。可能导致非关键业务的负载过高时，影响关键业务，主要是 restful 业务。计划在接下来的优化中，摘出高负载业务，使用独立 pod 或 serverless 架构进行部署，并配合任务队列，来实现流量削峰。`
- [ ] 增加群发模板消息功能
    1. 申请话术 `在考试公告发布后，使用模板消息向主动订阅的学员发送考试公告提醒。`
    1. 模板消息编辑后，需要有权限的用户审核后，才可以开始群发操作。审核权限在地市以上级别的管理员手中。
- [ ] 优化个人后缀业务
    - [ ] 向 OCC 平台申请获取个人推广编码详情接口 ( 获取推广地区(所属分部)、序列、姓名 )
    - [ ] 向 OCC 平台申请获取映射信息接口 ( 获取推广地区(所属分部)对应的 CRM 组织 ID、推广人 CRM 账号 ID )
    - [ ] 结合以上两个接口, 替代现有的需要自行维护的 CRM 组织信息业务、后缀花名册业务 `目前自行维护的方式, 容易出现后缀信息与 OCC 系统不匹配的情况；同时新增或修改推广人信息需要在多个平台重复操作，流程复杂，既不易于操作，也增加了工作量`

### 内部服务接口 Services
- [ ] 微信开放平台接口组
    - [ ] 授权事件接口
    - [ ] 授权方的消息与事件接口

### 管理平台接口 Manages
- [x] 认证
    - [x] 获取 RSA 公钥接口  
    `RSA 公钥改为从接口获取, 便于更新密钥对并增强安全性`
    - [x] 登陆接口  
    `登陆后返回会话 Token, 前端页面将 Token 保存在 localStorage 中, 在登陆时可以选择是否保持会话永不过期(通过在保存时增加过期时间字段实现), 减少登陆操作的次数，优化使用体验。`
        - [x] 登陆时增加校验 MisToken 的步骤，优化操作体验
    - [x] 退出登录 ( 销毁会话 ) 接口
    `使用 DeletedAt 字段进行软删除标注`
    - [x] 更新 MisToken 接口
    `前端的更新操作改为带遮罩且不可关闭的弹窗，不重置用户当前的操作，改进用户体验`
    - [x] 修改用户密码
    - [x] 获取用户基本信息
- [ ] 系统管理
    - [x] 配置管理
        - [x] 分页获取配置列表
        - [x] 修改配置
    - [x] 用户与角色管理
        - [x] 添加角色
        - [x] 获取当前用户所属角色及下属角色树
        - [x] 修改角色
        - [x] 修改角色的上级角色 ( 为前端的拖动修改角色归属功能提供服务 )
        - [x] 添加用户
        - [x] 分页获取用户列表
        - [x] 修改用户
        - [x] 禁用用户
        - [x] 启用用户
        - [x] 搜索用户

- [x] 业务接口校验到无对应的权限时，返回固定地无权限提示 `前端接收到无权限提示后，更新权限信息然后重定向路由到首页。达成权限变更后立刻体现的效果。`
- [x] 重新梳理权限分级的逻辑
  
- [ ] 摘出高资源消耗及高时间消耗的业务, 迁移到独立 pod 或 serverless 架构
    1. CRM 数据自动录入功能，摘出作为独立项目，由现在的长期运行，靠 sleep 实现定时轮询任务。修改为任务计划定期拉起 pod，这样可以在闲时节约资源。还可以测试迁移到 Serverless 的可能性。
    1. 同步粉丝列表功能, 从现在的收到请求就启动协程开始多线程处理 (无法预知同时进行的任务数量，可能导致某一个时刻多个任务在多线程处理，导致数据库压力极大)，修改为定时查询任务列表，逐个进行多线程处理
    1. 个人后缀在新增时，检查目前生效的所有后缀中，是否存在重复的。 后缀字段设为主键。 测试插入重复后缀时是否有效。
    1. 单点登陆模块进行迭代，去除活动名称字段，增加活动链接(多个链接用，分割)、活动表单ID
        1. 模块增加质检扣分

- [ ] 个人后缀相关服务
    - [ ] CRM 推送 ( 白皮书 ) ( 工作节点进行重构, 使用 CronJob 实现 )
    - [ ] 白皮书个人码批量生成 ( 获取中公营销平台访问令牌部分进行重构, 对获取到的访问令牌进行缓存, 取消现有的每次操作都获取一次的逻辑。 因为中公营销平台限制了只有最后一个令牌有效，现有的获取方式可能会出现多人同时操作时, 后者的操作会导致前者的令牌失效。 )
    - [ ] 宣传物料管理 ( 生成服务摘出, 使用 Serverless Frame 架构实现。 因为生成服务需要大量小号 CPU 资源及网络带宽资源, 现有架构会出现生成服务与关键业务争抢资源的情况。 )

### 外部活动接口 Events
- [ ] 海报生成业务迁移到 Serverless 架构 `需要为其提供瞬时的高算力与高带宽`

## 单元测试

`参考 https://www.jianshu.com/p/e3b2b1194830`

1. 安装 GoConvey
    1. 在命令行执行  
        `go get github.com/smartystreets/goconvey`
    1. 运行时间较长，运行完后  
        1. 在 $GOPATH/src 目录下新增了 github.com 子目录，该子目录里包含了 GoConvey 框架的库代码
        1. 在 $GOPATH/bin 目录下新增了 GoConvey 框架的可执行程序 GoConvey
    
1. 运行 Web 界面 并 在 Web 界面进行自动化编译测试
    1. 在项目根目录执行  
       `$GOPATH/bin/goconvey`

1. 可以参考的内容
    1. GoConvey 常用断言及用法 ( GoConvey断言err和bool的方法 ) https://blog.csdn.net/luomoshusheng/article/details/50226257?locationNum=13&fps=1
