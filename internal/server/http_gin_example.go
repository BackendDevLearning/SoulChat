package server

// 这是一个示例文件，展示如何在 Kratos HTTP Server 中集成 Gin
// 如果不需要 Gin，可以删除此文件

import (
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// GinHandler 将 Gin Engine 转换为 http.Handler
func GinHandler(ginEngine *gin.Engine) http.Handler {
	return ginEngine
}

// NewHTTPServerWithGin 示例：在 Kratos HTTP Server 中集成 Gin
func NewHTTPServerWithGin(c *conf.Server, jwtc *conf.JWT, s *service.ConduitService, logger log.Logger) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.ErrorEncoder(errorEncoder),
	}

	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratoshttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := kratoshttp.NewServer(opts...)

	// 创建 Gin Engine
	gin.SetMode(gin.ReleaseMode)
	ginEngine := gin.New()
	ginEngine.Use(gin.Logger(), gin.Recovery())

	// 在 Gin 中定义路由
	ginEngine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 将 Gin Engine 挂载到 Kratos HTTP Server
	// 方式1：挂载到特定路径前缀
	srv.HandlePrefix("/api/v1/", GinHandler(ginEngine))

	// 方式2：挂载到根路径（会覆盖 Kratos 的路由）
	// srv.Handle("/", GinHandler(ginEngine))

	// 仍然可以使用 Kratos 的 protobuf 路由
	// v1.RegisterConduitHTTPServer(srv, s)

	return srv
}

