package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v4"
)

var currentUserKey struct{}

type CurrentUser struct {
	UserID uint
}

// GenerateToken JWT格式: header.payload.signature
func GenerateToken(secret string, userid uint32, expire time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid": userid,
		// 开发阶段直接写死
		"nbf": time.Date(2000, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		// 实际token添加过期时间
		//"iat": time.Now().Unix(),             // 签发时间 Issued At
		//"nbf": time.Now().Unix(),             // 生效时间 Not Before
		//"exp": time.Now().Add(expire).Unix(), // 过期时间 Expiration Time
	})

	// 根据secret生成最终token string
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func JWTAuth(secret string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				tokenString := tr.RequestHeader().Get("Authorization")
				auths := strings.SplitN(tokenString, " ", 2)
				if len(auths) != 2 || !strings.EqualFold(auths[0], "Token") {
					return nil, fmt.Errorf("jwt token missing")
				}

				token, err := jwt.Parse(auths[1], func(token *jwt.Token) (interface{}, error) {
					// Don't forget to validate the alg is what you expect:
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(secret), nil
				})

				if err != nil {
					return nil, err
				}

				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					// put CurrentUser into ctx
					if u, ok := claims["userid"]; ok {
						ctx = WithContext(ctx, &CurrentUser{UserID: uint(u.(float64))})
					}
				} else {
					return nil, fmt.Errorf("token invalid")
				}
			}
			return handler(ctx, req)
		}
	}
}

func FromContext(ctx context.Context) *CurrentUser {
	return ctx.Value(currentUserKey).(*CurrentUser)
}

func WithContext(ctx context.Context, user *CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}
