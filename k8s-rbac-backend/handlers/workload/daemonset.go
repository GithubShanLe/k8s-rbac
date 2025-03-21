package workload

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListDaemonsetRequest struct {
	DaemonsetName string `json:"daemonsetName"`
	NameSpace     string `json:"namespace"`
}

type ListDaemonsetResponse struct {
	handlers.ErrorResponse
	Daemonsets []Daemonset `json:"daemonsets"`
}

type Daemonset struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Labels     map[string]string `json:"labels"`
	Images     []string          `json:"images"`
	Pods       string            `json:"pods"`
	CreateTime string            `json:"createTime"`
}

func ListDaemonset(w http.ResponseWriter, r *http.Request) {
	clientset := k8s.GetClient()
	var resp ListDaemonsetResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()

	// 解析请求参数
	var req ListDaemonsetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	listOptions := metav1.ListOptions{}
	if req.DaemonsetName != "" {
		listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", req.DaemonsetName)
	}
	svcs, err := clientset.AppsV1().DaemonSets(req.NameSpace).List(context.Background(), listOptions)
	if err != nil {
		resp.ErrorCode = "500"
		resp.ErrorMessage = fmt.Sprintf("获取svc列表失败: %v", err)
		return
	}
	for _, svc := range svcs.Items {
		resp.Daemonsets = append(resp.Daemonsets, Daemonset{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    svc.Labels,
			Images: func() []string {
				var images []string
				for _, container := range svc.Spec.Template.Spec.Containers {
					images = append(images, container.Image)
				}
				return images
			}(),
			Pods:       fmt.Sprintf("%d/%d", svc.Status.CurrentNumberScheduled, svc.Status.DesiredNumberScheduled),
			CreateTime: svc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	return
}
