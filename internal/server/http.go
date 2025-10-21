package server

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"
	"kratos-realworld/internal/service"
	swaggerui "kratos-realworld/internal/swagger-ui"
	wsrv "kratos-realworld/internal/websocket"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
)

type websocketHandler struct {
	jwtc *conf.JWT
}

func NewWebsocketHandler(jwtc *conf.JWT) *websocketHandler {
	return &websocketHandler{
		jwtc: jwtc,
	}
}

func (h *websocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 将WebSocket服务也统一使用JWT验证，通过token取得userID
	// 从header或query中获取token
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		tokenString = r.URL.Query().Get("token")
	}

	if tokenString == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
	}

	// 去掉Token前缀
	auths := strings.SplitN(tokenString, " ", 2)
	if len(auths) == 2 && strings.EqualFold(auths[0], "Token") {
		tokenString = auths[1]
	}

	// 解析token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtc.Secret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
	}

	// 取得userID
	userID := uint32(claims["userid"].(float64))
	log.Debug("[WS] 用户 %d 成功通过JWT鉴权 ", userID)

	// 允许跨域升级
	var upGrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug("WebSocket upgrade error: ", err)
		return
	}

	// 这里的 Client、MyServer 来自 internal/websocket 包
	c := &wsrv.Client{
		Name: strconv.Itoa(int(userID)),
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

	// s.mc 是 ConduitService 里已经初始化的 MessageUseCase
	wsrv.InitWebsocketServer(s.GetMessageUseCase())
	srv.Handle("/ws", NewWebsocketHandler(jwtc))

	return srv
}
