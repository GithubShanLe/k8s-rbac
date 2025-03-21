package main

import (
	"fmt"
	"log"
	"net/http"

	"k8s-rbac-backend/handlers"
	nodepool "k8s-rbac-backend/handlers/node_pool"
	"k8s-rbac-backend/handlers/sa"
	"k8s-rbac-backend/handlers/service"
	"k8s-rbac-backend/handlers/workload"
	"k8s-rbac-backend/middleware"
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
		case "/api/node/list":
			nodepool.ListClusterNodes(w, r)
		case "/api/svc/list":
			service.ListService(w, r)
		case "/api/workload/deployment/list":
			workload.ListDeployment(w, r)
		case "/api/workload/replicaset/list":
			workload.ListReplicaset(w, r)
		case "/api/workload/pod/list":
			workload.ListPod(w, r)
		case "/api/workload/job/list":
			workload.ListJob(w, r)
		case "/api/workload/cronjob/list":
			workload.ListCronJob(w, r)
		case "/api/workload/daemonset/list":
			workload.ListDaemonset(w, r)
		case "/api/workload/statefulset/list":
			workload.Liststatefulset(w, r)
		case "/api/workload/pod/metrics":
			workload.GetPodMetric(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// 应用中间件
	handler := middleware.HandleAllNamespace(apiHandler)

	// 应用跨域中间件并注册到路由复用器
	mux.Handle("/api/", enableCORS(handler))

	// 启动 HTTP 服务器
	fmt.Println("服务器启动，监听端口 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
