# P1-01 · Redis 与 Context 常量表

> 复刻目标：`szluyu99/gin-vue-blog` 的 `gin-blog-server`
> 本节交付：`internal/global/keys.go`

---

## 1.1 这一节解决什么问题

P1 开始要接 Redis 和数据库了，随之而来的是一堆**字符串键名**：

```go
rdb.Get(ctx, "view_count")
rdb.HGetAll(ctx, "article_like_count")
c.MustGet("_db_field")
```

这些字符串如果散落在 30 个文件里，会出现三类问题：

1. **拼错不会报错**。`"view_count"` 写成 `"view_conut"`，编译通过，运行时静默读到空值 —— 最恶心的一类 bug。
2. **改不了**。想把 `visitor_area` 改成 `visitor:area`，你得全项目搜索替换，还得担心误伤。
3. **看不出全貌**。新人想知道「这个项目用了 Redis 存什么」，得把整个项目翻一遍。

`keys.go` 的作用就是**把这些魔法字符串收进一个文件，全部变成常量**。

---

## 1.2 完整代码

新建 `internal/global/keys.go`：

```go
package g

// Redis Key

const (
	// MAIL_CODE = "mail_code:" // 验证码
	// DELETE       = "delete:"      //? 记录强制下线用户?
	ONLINE_USER  = "online_user:"  // 在线用户
	OFFLINE_USER = "offline_user:" // 强制下线用户
	LOGIN_FAIL   = "login_fail:"   // 登录连续失败次数 (login_fail:<用户名+IP 的哈希>), 带 TTL
	VISITOR_AREA = "visitor_area"  // 地域统计
	VIEW_COUNT   = "view_count"    // 访问数量

	/*
		按天的访问量 (view_count:2006-01-02), 带 30 天 TTL

		和 VIEW_COUNT 不是一个口径: 那个是历史独立访客累计(同一访客只算一次),
		这个是当天的访问次数(每次打开前台算一次), 所以按天的加起来不等于累计值。
		只用来画趋势图, 过期即丢 —— 数据库里没有对应字段, 不必长期保留。
	*/
	VIEW_COUNT_DAY = "view_count:"

	// 前端错误上报的按 IP 配额 (error_report:<IP>), 带 TTL。
	// 上报接口匿名可访问, 没有配额等于开了个无鉴权的写库入口
	ERROR_REPORT = "error_report:"

	KEY_UNIQUE_VISITOR_SET = "unique_visitor" // 唯一用户记录 set

	ARTICLE_USER_LIKE_SET = "article_user_like:" // 文章点赞 Set
	ARTICLE_LIKE_COUNT    = "article_like_count" // 文章点赞数
	ARTICLE_VIEW_COUNT    = "article_view_count" // 文章查看数
	// 文章浏览去重, 格式 article_view_visitor:<文章id>:<访客指纹>, 带 TTL
	ARTICLE_VIEW_VISITOR = "article_view_visitor:"

	COMMENT_USER_LIKE_SET = "comment_user_like:" // 评论点赞 Set
	COMMENT_LIKE_COUNT    = "comment_like_count" // 评论点赞数

	PAGE   = "page"   // 页面封面
	CONFIG = "config" // 博客配置
)

// Gin Context Key | Session Key

const (
	CTX_DB        = "_db_field"
	CTX_RDB       = "_rdb_field"
	CTX_USER_AUTH = "_user_auth_field"
)

// Config Key

const (
	CONFIG_ARTICLE_COVER     = "article_cover"
	CONFIG_IS_COMMENT_REVIEW = "is_comment_review"
	CONFIG_IS_MESSAGE_REVIEW = "is_message_review"
	CONFIG_ABOUT             = "about"
)
```

---

## 1.3 逐段讲解

### ① 为什么这个文件不用 import 任何东西

注意文件头：**没有 import 块**。

这是它作为「零依赖文件」的标志。它只声明常量，不引用任何包。这种文件在项目里是最好的东西 —— 可以被任何层引用，自己不会引入任何耦合。

对比一下：`config.go` 引入了 viper，`result.go` 引入了 fmt，这个什么都不引。

### ② 三类常量，三种用途

文件用三段注释分了三个组：

| 组 | 用途 | 谁用 |
|---|---|---|
| Redis Key | Redis 里键的名字 | handle / counter / middleware |
| Gin Context Key | 往 `gin.Context` 里存东西时的 key | middleware 存，handle 取 |
| Config Key | `CONFIG` 那个 Hash 里的字段名 | handle_bloginfo |

**关键：第二组和第一组是不同层面的东西，别混。**

- **Redis Key** 是存在 Redis 服务器里的字符串（跨进程、跨机器都能看到）
- **Context Key** 是存在**当前这次 HTTP 请求的内存对象**里的（请求结束就没了）

它们放在同一个文件里，是因为都扮演同一个角色 —— 「防止字符串写错的常量表」。

### ③ Context Key 为什么长这样：`"_db_field"`

```go
CTX_DB        = "_db_field"
CTX_RDB       = "_rdb_field"
CTX_USER_AUTH = "_user_auth_field"
```

**为什么加下划线前缀和后缀？**

理论上你可以直接写：

```go
CTX_DB = "db"      // ❌ 危险
```

隐患在于**gin 自己也往 Context 里塞东西**，用的 key 有 `"keys"`、`"gin_context"` 等。更麻烦的是**第三方中间件**（你自己以后加的）可能用 `"db"`、`"user"` 这种直觉名字。

一旦撞了：

```go
c.Set("db", gormDB)      // 你塞的
c.Set("db", redisClient) // 某个第三方中间件覆盖了
db := c.MustGet("db")    // 类型断言失败 → panic
```

`"_db_field"` 这种带下划线前后缀的命名，是**故意起得丑**，让撞名概率趋近于零。这是防御性命名。

### ④ 键名末尾的冒号 —— 一个必须看懂的约定

```go
ONLINE_USER  = "online_user:"     // ← 有冒号
VISITOR_AREA = "visitor_area"     // ← 没冒号
VIEW_COUNT   = "view_count"       // ← 没冒号
VIEW_COUNT_DAY = "view_count:"    // ← 有冒号
```

**规律：末尾带冒号的，是「前缀」，后面还要拼东西。**

```go
onlineKey := g.ONLINE_USER + strconv.Itoa(auth.ID)
//          "online_user:" + "5"  →  "online_user:5"
```

不带冒号的，是**完整的键名**：

```go
rdb.Get(ctx, g.VIEW_COUNT)        // 直接就是 "view_count"
```

**如果不加冒号会怎样？**

```go
ONLINE_USER = "online_user"       // 少个冒号
onlineKey := g.ONLINE_USER + "5"  // → "online_user5"
```

也能跑，但两个问题：

- 可读性差：Redis 里看到 `online_user5` 不知道 5 是什么
- **前缀歧义**：如果还有个键叫 `online_user_role`，用 `KEYS online_user*` 扫描时会把它一起扫出来。有冒号分隔就清晰得多

Redis 社区惯例就是用冒号做层级分隔（`user:5:profile`），这个项目沿用了。

### ⑤ 注释里的三段「为什么」，值得单独读

原版在这些常量上写的注释，信息量很大：

**第一段 —— 两个 VIEW_COUNT 不是一个口径：**

> `VIEW_COUNT` 是历史独立访客累计（同一访客只算一次），`VIEW_COUNT_DAY` 是当天的访问次数（每次打开前台算一次），所以按天的加起来不等于累计值。

**这是个很容易搞混的地方**：看到两个都叫「访问量」的键，直觉以为一个是总计、一个是明细。实际上它们的**去重口径不同** —— 一个是「有多少人来过」，一个是「今天被打开了多少次」。

**第二段 —— ERROR_REPORT 是安全考虑：**

> 上报接口匿名可访问，没有配额等于开了个无鉴权的写库入口

前端错误上报接口必须匿名（访客还没登录就可能出错），但它会**写数据库**。这就等于开了个「谁都能写」的入口，必须靠按 IP 的配额兜住。

**第三段 —— 点赞用了两个键：**

```go
ARTICLE_USER_LIKE_SET = "article_user_like:" // Set: 谁点过
ARTICLE_LIKE_COUNT    = "article_like_count" // Hash: 点了几次
```

一个 Set 记「谁点过赞」（用于判断「我点过没有」，防重复点赞），一个 Hash 记「总点赞数」（用于列表展示）。**两个键职责不同，不是冗余。**

> 这段设计在 P7 会详细展开。现在只要建立印象：**Redis 里同一件事常常需要两种组织方式，一种服务于「查询」，一种服务于「展示」。**

---

## 1.4 验收

```bash
cd D:/code/gocode/blog/MyBlog/gin-blog-server
go build ./...
```

**期望**：无输出。

这个文件没有任何依赖，编译不过只可能是复制粘贴漏了括号。
