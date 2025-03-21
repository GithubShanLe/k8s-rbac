package handlers

import (
	"context"
	"encoding/json"
	"k8s-rbac-backend/k8s"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNamespaces 获取所有命名空间

type GetNamespaceResponse struct {
	Namespaces []string `json:"namespaces"`
}

func GetNamespaces(w http.ResponseWriter, r *http.Request) {
	clientset := k8s.GetClient()

	// 获取命名空间列表
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp = new(GetNamespaceResponse)
	for _, item := range namespaces.Items {
		resp.Namespaces = append(resp.Namespaces, item.Name)
	}
	//用于前段显示，表示所有命名空间
	resp.Namespaces = append(resp.Namespaces, "all+")
	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
