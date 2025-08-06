package swagger_ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var swaggerFiles embed.FS

func Handler() http.Handler {
	// 把 dist/ 子目录作为根路径暴露出来
	distFS, err := fs.Sub(swaggerFiles, "dist")
	if err != nil {
		panic(err)
	}
	return http.StripPrefix("/swagger-ui/", http.FileServer(http.FS(distFS)))
}

//go:embed openapi.yaml
var openapiFile embed.FS

func HandlerOpenapi() http.Handler {
	return http.FileServer(http.FS(openapiFile))
}
