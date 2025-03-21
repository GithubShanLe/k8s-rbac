package nodepool

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListNodeRequest struct {
	NodeName string `json:"nodeName"`
}

type ListNodeResponse struct {
	handlers.ErrorResponse
	Nodes []Node `json:"nodes"`
}

func ListClusterNodes(w http.ResponseWriter, r *http.Request) {
	clientset := k8s.GetClient()
	var resp ListNodeResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()

	// 解析请求参数
	var req ListNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	// 构建 ListOptions
	// 构建筛选条件，支持按节点名称或IP筛选
	var fieldSelector string
	if req.NodeName != "" {
		fieldSelector = fmt.Sprintf("metadata.name=%s", req.NodeName)
	}
	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}
	// 正确顺序：先获取节点列表
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		resp.ErrorCode = "500"
		resp.ErrorMessage = fmt.Sprintf("获取节点列表失败: %v", err)
		return
	}

	// 处理节点数据
	nodeList := make([]Node, 0)
	for _, node := range nodes.Items {
		nodeList = append(nodeList, getNodes(node))
	}
	resp.Nodes = nodeList
}
