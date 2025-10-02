package server

import (
	"context"
	"fmt"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"
	"kratos-realworld/internal/service"
	wsrv "kratos-realworld/internal/websocket"
	swaggerui "kratos-realworld/internal/swagger-ui"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
)

type websocketHandler struct{}

func (h *websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 允许跨域升级
	var upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("websocket upgrade error:", err)
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		fmt.Println("websocket missing 'user' query parameter")
		_ = conn.Close()
		return
	}

	// 这里的 Client、MyServer 来自 internal/websocket 包
	c := &wsrv.Client{
		Name: user,
		Conn: conn,
		Send: make(chan []byte, 16),
	}
	wsrv.MyServer.Register <- c
	go c.Read()
	go c.Write()
}



func NewSkipRoutersMatcher() selector.MatchFunc {

	skipRouters := map[string]struct{}{
		"/realworld.v1.Conduit/Login":        {},
		"/realworld.v1.Conduit/Register":     {},
		"/realworld.v1.Conduit/GetArticle":   {},
		"/realworld.v1.Conduit/ListArticles": {},
		"/realworld.v1.Conduit/GetComments":  {},
		"/realworld.v1.Conduit/GetTags":      {},
		"/realworld.v1.Conduit/GetProfile":   {},
	}

	return func(ctx context.Context, operation string) bool {
		if _, ok := skipRouters[operation]; ok {
			return false
		}
		return true
	}
}

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, jwtc *conf.JWT, s *service.ConduitService, logger log.Logger) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.ErrorEncoder(errorEncoder),

		kratoshttp.Middleware(
			recovery.Recovery(),
			selector.Server(auth.JWTAuth(jwtc.Secret)).Match(NewSkipRoutersMatcher()).Build(),
			logging.Server(logger),
		),
		kratoshttp.Filter(
			handlers.CORS(
				handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
				handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"}),
				handlers.AllowedOrigins([]string{"*"}),
			),
		),
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
	v1.RegisterConduitHTTPServer(srv, s)

	// 注册 Swagger UI
	srv.Handle("/openapi.yaml", swaggerui.HandlerOpenapi())
	srv.HandlePrefix("/swagger-ui/", swaggerui.Handler())

	srv.Handle("/ws", &websocketHandler{})

	return srv
}
