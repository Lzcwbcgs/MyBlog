# 从零复刻 gin-blog-server · 施工路线

对照物：`D:\code\gocode\blog\gin-vue-blog\gin-blog-server\`（原版，只读，一行不动）
施工地：`D:\code\gocode\blog\MyBlog\gin-blog-server\`（本文件所在目录）

## 复刻原则

1. **文件路径、包名、函数名、结构体名全部对齐原版**，以便随时逐文件 diff。
2. **连坑一起复刻**。原版注释里标注过的取舍（model 层职责偏重、响应信封吃掉 HTTP 语义等）原样保留，写的时候讲清楚为什么。
3. **每个阶段结束时必须能跑**。交付物 = 可运行的服务 + 一条 curl 或 `go test` 验收命令。
4. 注释写「为什么」，不写「是什么」—— 这是原版最值得学的部分。

## 阶段总览

| 阶段 | 主题 | 交付 | 验收 |
|---|---|---|---|
| P0 | 骨架 + 配置 + 第一个接口 | `go.mod` / `config.yml` / `global/config.go` / `global/result.go` / `cmd/main.go` | `curl /api/ping` 返回信封 |
| P1 | 基础设施 + 中间件底座 | `helper.go` / `global/keys.go` / `handle/base.go` / `middleware/base.go` | 服务连上 DB+Redis，请求走完整中间件链 |
| P2 | 数据层 + 第一个 CRUD | `model/z_base.go` / `model/category.go` / `model/tag.go` / 两个 handle / 路由注册 | curl 分类列表返回分页数据 |
| P3 | **认证与权限（核心）** | `utils/jwt` / `utils/encrypt` / `model/auth.go` / `model/auth_control.go` / `middleware/auth.go` / `handle_auth.go` / `model/seed_resource.go` / `cmd/generate-data` | 登录拿 token；无 token 被拦；越权被拦 |
| P4 | 日志中间件 + IP 工具 | `utils/ip.go` / `middleware/operation_log.go` / `middleware/listen_online.go` / 三个日志 model | POST 后库里有记录，密码被脱敏 |
| P5 | 内容模块（广度） | article / comment / talk / message / friend_link / page 的 model + handle | 后台六类内容增删改查全通 |
| P6 | 前台接口 | `handle_front.go` / `handle_bloginfo.go` / `/api/front` 路由组 | 前台首页 / 列表 / 详情 / 评论 / 归档 / 搜索可用 |
| P7 | Redis 计数与缓存 | `handle/cache.go` / `counter.go` / `model/counter_snapshot.go` | 点赞后重启服务，计数不归零 |
| P8 | 通知 + 错误上报 + 上传 | notification / error_log / upload 三块 | 评论触发通知；上报限流生效；图片可访问 |
| P9 | 测试补齐 + Swagger + 部署 | 8 个包的测试 / swag 注解 / Dockerfile / compose | `go test ./...` 全绿 |

## 每阶段的固定节奏

1. 我讲这一层要解决什么问题、关键决策点、容易踩的坑
2. 我给**接口契约**（文件名 / 包名 / 函数签名 / 结构体字段）—— 必须对齐原版，否则没法 diff
3. **你动手敲**，卡住了随时问
4. 我逐行 review，对着原版 diff，讲差异

## 阶段依赖图

```
P0 ──→ P1 ──→ P2 ──→ P3 ──┬──→ P4 ──→ P5 ──→ P6 ──→ P7 ──→ P8 ──→ P9
                          └──→ P5 (P4 与 P5 可并行)
```

P3 是分水岭：它之前是「一个能跑的 Go Web 服务」，它之后才是「这个项目」。

## 环境

- Go 1.26.4
- Redis 已在 `127.0.0.1:6379` 运行（原版 dev.sh 用的是 DB 7）
- 数据库默认 SQLite（`gvb.db`），不需要 MySQL
- 原版配置里 `Server.Port` 是 `:8765`，两边同时跑不冲突
