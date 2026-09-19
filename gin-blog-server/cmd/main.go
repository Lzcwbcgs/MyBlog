package main

import (
	"flag"
	"net/http"

	g "gin-blog/internal/global"

	"github.com/gin-gonic/gin"
)

func main() {
	// 配置文件路径, 默认是 ../config.yml
	// 注意: 这是相对【工作目录】解析的, 所以必须在 cmd/ 下运行, 见 4.3 ⑧
	configPath := flag.String("c", "../config.yml", "配置文件路径")
	flag.Parse()

	// 根据命令行参数读取配置文件, 其他变量的初始化依赖于配置文件对象
	conf := g.ReadConfig(*configPath)

	// 只信任配置里的代理网段
	// 之前是 SetTrustedProxies("*"), 访客随手加个 X-Real-IP 就能伪造来源 IP
	gin.SetMode(conf.Server.Mode)
	r := gin.New()

	// 开发模式使用 gin 自带的日志和恢复中间件, 生产模式使用自定义的 (P1 才会有)
	if conf.Server.Mode == "debug" {
		r.Use(gin.Logger(), gin.Recovery()) // gin 自带的日志和恢复中间件, 挺好用的
	} else {
		// TODO(P1): 换成 middleware.Recovery(true), middleware.Logger()
		r.Use(gin.Recovery())
	}

	if err := r.SetTrustedProxies(conf.TrustedProxies()); err != nil {
		panic("可信代理网段配置有误: " + err.Error())
	}

	// P0 的验收接口: P1 会把它移进路由注册表
	r.GET("/api/ping", func(c *gin.Context) {
		// 这里手工拼信封是临时的, P1 会有 handle.Response[T] 和 ReturnSuccess
		c.JSON(http.StatusOK, gin.H{
			"code":    g.OkResult.Code(),
			"message": g.OkResult.Msg(),
			"data":    "pong",
		})
	})

	if err := r.Run(conf.Server.Port); err != nil {
		panic("HTTP 服务启动失败: " + err.Error())
	}
}
