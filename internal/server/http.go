package server

import (
	"context"
	"fmt"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"
	"kratos-realworld/internal/service"
	swaggerui "kratos-realworld/internal/swagger-ui"
	wsrv "kratos-realworld/internal/websocket"
	"net/http"
	"strconv"
	"strings"

	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"kratos-realworld/internal/chat"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	// 检查连接的Origin头安全检查函数，用于验证请求的 Origin 头，防止跨站 WebSocket 劫持（CSWSH）。
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsLoginHandler struct {
	jwtc  *conf.JWT
	chatU *chat.ChatUsecase
}

func NewWsLoginHandler(jwtc *conf.JWT, chatU *chat.ChatUsecase) *wsLoginHandler {
	return &wsLoginHandler{
		jwtc:  jwtc,
		chatU: chatU,
	}
}

type wsLogoutHandler struct {
	jwtc  *conf.JWT
	chatU *chat.ChatUsecase
}

func NewWsLogoutHandler(jwtc *conf.JWT, chatU *chat.ChatUsecase) *wsLogoutHandler {
	return &wsLogoutHandler{
		jwtc:  jwtc,
		chatU: chatU,
	}
}

func (h *wsLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	log.Debugf("[WS] 用户 %d 成功通过JWT鉴权 ", userID)

	// 在 http 层做 Upgrade，然后把已升级的 conn 传给用例层处理
	conn, err := upgrader.Upgrade(w, r, nil)

	if h.chatU == nil {
		http.Error(w, "chat usecase not configured", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "failed to upgrade websocket: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// 交给用例层处理连接（非阻塞）
	go h.chatU.Login(conn, strconv.Itoa(int(userID)))
}

func (h *wsLogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 暂不使用前端传入提交的 client_id；server 使用 JWT 中的 userid 作为 client id
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		tokenString = r.URL.Query().Get("token")
	}

	if tokenString == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
	}

	auths := strings.SplitN(tokenString, " ", 2)
	if len(auths) == 2 && strings.EqualFold(auths[0], "Token") {
		tokenString = auths[1]
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtc.Secret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// 从 token 中取出 userid，作为要登出的 client id
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
		return
	}
	userID := uint32(claims["userid"].(float64))
	clientId := strconv.Itoa(int(userID))

	// 调用用例层的 Logout
	msg, code := h.chatU.Logout(clientId)
	var errVal error
	if code != 0 {
		errVal = fmt.Errorf(msg)
	}
	resp := map[string]interface{}{"code": code, "res": service.ErrorToRes(errVal)}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func NewSkipRoutersMatcher() selector.MatchFunc {

	skipRouters := map[string]struct{}{
		"/realworld.v1.Conduit/Register":   {},
		"/realworld.v1.Conduit/Login":      {},
		"/realworld.v1.Conduit/LoginBySms": {},
		"/realworld.v1.Conduit/SendSms":    {},
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

	// 注册 websocket login/logout 的 HTTP 接口，login 做 upgrade 并交给 chat 用例处理
	srv.Handle("/api/ws/login", NewWsLoginHandler(jwtc, s.GetChatUsecase()))
	srv.Handle("/api/ws/logout", NewWsLogoutHandler(jwtc, s.GetChatUsecase()))

	// s.mc 是 ConduitService 里已经初始化的 MessageUseCase
	wsrv.InitWebsocketServer(s.GetMessageUseCase())

	return srv
}
