package nodepool

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListNodePoolRequest struct {
	NodePoolName string `json:"nodePoolName"`
}

type ListNodePoolResponse struct {
	handlers.ErrorResponse
	NodePools []NodePool `json:"nodePools"`
}
type NodePool struct {
	Name     string            `json:"name"`
	Status   string            `json:"status"` //暂时不使用
	Lables   map[string]string `json:"lables"`
	Taints   map[string]string `json:"taints"`
	NodeList []Node            `json:"nodeList"`
}

type Node struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	RequestCpu int64  `json:"requestCpu"`
	LimitCpu   int64  `json:"limitCpu"`
	RequestMem int64  `json:"requestMem"`
	LimitMem   int64  `json:"limitMem"`
	CurrentPod int64  `json:"currentPod"`
	RequestPod int64  `json:"requestPod"`
	LimitPod   int64  `json:"limitPod"`
	NodeIp     string `json:"nodeIp"`
	CreatedAt  string `json:"createdAt"`
}

// 获取节点池和节点列表
func ListNodePool(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var resp ListNodePoolResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()
	var req ListNodePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	// 获取 Kubernetes 客户端
	clientset := k8s.GetClient()

	// 统一构造 ListOptions
	labelSelector := ""
	if req.NodePoolName != "" {
		labelSelector = fmt.Sprintf("nodepool=%s", req.NodePoolName)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		resp.ErrorCode = "500"
		resp.ErrorMessage = fmt.Sprintf("获取节点列表失败: %v", err)
		return
	}

	// 按标签分组节点
	nodePoolMap := make(map[string]*NodePool)
	for _, node := range nodes.Items {
		pod, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
		})
		if err != nil {
			resp.ErrorCode = "500"
			resp.ErrorMessage = fmt.Sprintf("获取节点Pod失败: %v", err)
			return
		}
		// 创建节点信息
		nodeInfo := Node{
			Name:       node.Name,
			Status:     GetNodeStatus(node),
			RequestCpu: node.Status.Allocatable.Cpu().MilliValue(),
			LimitCpu:   node.Status.Capacity.Cpu().MilliValue(),
			RequestMem: node.Status.Allocatable.Memory().MilliValue() / (1024 * 1024), // 转换为 MB
			LimitMem:   node.Status.Capacity.Memory().MilliValue() / (1024 * 1024),
			CurrentPod: int64(len(pod.Items)),
			RequestPod: node.Status.Allocatable.Pods().Value(),
			LimitPod:   node.Status.Capacity.Pods().Value(),
			NodeIp:     node.Status.Addresses[0].Address,
			CreatedAt:  node.CreationTimestamp.Format("2006-01-02 15:04:05"),
		}

		// 获取节点标签
		labels := node.Labels
		nodePoolKey := "default"
		fmt.Println(labels)
		for l, v := range labels {
			if strings.Contains(l, "nodepool") || strings.Contains(l, "nodegroup") {
				nodePoolKey = v
			}
		}

		// 如果节点池不存在，创建新的节点池
		if _, exists := nodePoolMap[nodePoolKey]; !exists {
			nodePoolMap[nodePoolKey] = &NodePool{
				Status:   "Active",
				Lables:   labels,
				Taints:   make(map[string]string),
				NodeList: make([]Node, 0),
				Name:     nodePoolKey,
			}
			// 添加污点信息
			for _, taint := range node.Spec.Taints {
				nodePoolMap[nodePoolKey].Taints[taint.Key] = string(taint.Effect)
			}
		}

		// 将节点添加到对应的节点池
		nodePoolMap[nodePoolKey].NodeList = append(nodePoolMap[nodePoolKey].NodeList, nodeInfo)
	}

	// 转换 map 为数组
	for _, pool := range nodePoolMap {
		resp.NodePools = append(resp.NodePools, *pool)
	}
	return
}
