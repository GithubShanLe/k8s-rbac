package main

import (
	"fmt"
	"log"
	"net/http"

	"k8s-rbac-backend/handlers"
	nodepool "k8s-rbac-backend/handlers/node_pool"
	"k8s-rbac-backend/handlers/sa"
)

// enableCORS 添加跨域支持的中间件
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 创建路由复用器
	mux := http.NewServeMux()

	// API 路由
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/resources":
			handlers.GetResources(w, r)
		case "/api/verbs":
			handlers.GetVerbs(w, r)
		case "/api/create-sa":
			sa.CreateServiceAccount(w, r)
		case "/api/listSa":
			sa.ListServiceAccounts(w, r)
		case "/api/sa-details":
			sa.GetServiceAccountDetails(w, r)
		case "/api/ns":
			handlers.GetNamespaces(w, r)
		case "/api/update-sa":
			sa.UpdateSa(w, r)
		case "/api/nodepool/list":
			nodepool.ListNodePool(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// 应用跨域中间件到 API 路由
	mux.Handle("/api/", enableCORS(apiHandler))

	// 启动 HTTP 服务器
	fmt.Println("服务器启动，监听端口 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
