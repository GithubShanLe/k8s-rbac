package service

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListServiceRequest struct {
	ServiceName string `json:"serviceName"`
	NameSpace   string `json:"namespace"`
}

type ListServiceResponse struct {
	handlers.ErrorResponse
	Services []Service `json:"services"`
}

type Service struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Type      string            `json:"type"`
	ClusterIp string            `json:"clusterIp"`
	LBIp      string            `json:"lbIp"`
	// 外部端点列表
	ExternalEndpoints []string `json:"externalEndpoints"`
	// 内部端点列表
	InternalEndpoints []string `json:"internalEndpoints"`

	CreateTime string `json:"createTime"`
}

func ListService(w http.ResponseWriter, r *http.Request) {
	clientset := k8s.GetClient()
	var resp ListServiceResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()

	// 解析请求参数
	var req ListServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	listOptions := metav1.ListOptions{}
	if req.ServiceName != "" {
		listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", req.ServiceName)
	}
	svcs, err := clientset.CoreV1().Services(req.NameSpace).List(context.Background(), listOptions)
	if err != nil {
		resp.ErrorCode = "500"
		resp.ErrorMessage = fmt.Sprintf("获取svc列表失败: %v", err)
		return
	}
	for _, svc := range svcs.Items {
		resp.Services = append(resp.Services, Service{
			Name:              svc.Name,
			Namespace:         svc.Namespace,
			Labels:            svc.Labels,
			Type:              string(svc.Spec.Type),
			ClusterIp:         svc.Spec.ClusterIP,
			LBIp:              svc.Spec.LoadBalancerIP,
			ExternalEndpoints: svc.Spec.ExternalIPs,
			InternalEndpoints: func() []string {
				var eps []string
				for _, ep := range svc.Spec.Ports {
					eps = append(eps, fmt.Sprintf("%s.%s:%d %s", svc.Name, svc.Namespace, ep.Port, ep.Protocol))
				}
				return eps
			}(),
			CreateTime: svc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	return
}
