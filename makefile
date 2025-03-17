# 定义变量
FRONTEND_DIR := k8s-rbac-web
BACKEND_DIR := k8s-rbac-backend
BINARY_NAME := k8s-rbac-backend

.PHONY: all clean build run deploy

# 默认目标
all: build

# 清理构建产物
clean:
	@echo "正在清理构建产物..."
	rm -rf $(FRONTEND_DIR)/dist
	rm -f $(BACKEND_DIR)/$(BINARY_NAME)

# 构建前端
build-frontend:
	@echo "正在构建前端..."
	cd $(FRONTEND_DIR) && npm install && npm run build

# 构建后端
build-backend:
	@echo "正在构建后端..."
	cd $(BACKEND_DIR) && go build -o $(BINARY_NAME)

# 构建所有
build: build-frontend build-backend

# 运行开发环境
dev-frontend:
	@echo "启动前端开发服务器..."
	cd $(FRONTEND_DIR) && npm run dev

dev-backend:
	@echo "启动后端开发服务器..."
	cd $(BACKEND_DIR) && go run main.go

# 部署前端
deploy-frontend:
	@echo "部署前端..."
	cd $(FRONTEND_DIR) && \
	docker build -t k8s-rbac-web:latest . && \
	kubectl apply -f k8s/frontend-deployment.yaml

# 部署后端
deploy-backend:
	@echo "部署后端..."
	cd $(BACKEND_DIR) && \
	docker build -t k8s-rbac-backend:latest . && \
	kubectl apply -f k8s/backend-deployment.yaml

# 部署所有
deploy: deploy-frontend deploy-backend

# 帮助信息
help:
	@echo "可用的命令："
	@echo "  make clean         - 清理构建产物"
	@echo "  make build        - 构建前后端"
	@echo "  make build-frontend - 仅构建前端"
	@echo "  make build-backend  - 仅构建后端"
	@echo "  make dev-frontend   - 启动前端开发服务器"
	@echo "  make dev-backend    - 启动后端开发服务器"
	@echo "  make deploy        - 部署前后端到 Kubernetes"
	@echo "  make deploy-frontend - 仅部署前端"
	@echo "  make deploy-backend  - 仅部署后端"