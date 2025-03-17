package sa

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateRoleRequest 定义更新 Role 的请求结构
type UpdateSaRequest struct {
	ServiceAccountName string  `json:"serviceAccountName"`
	RoleName           string  `json:"roleName"`
	ClusterRoleName    string  `json:"clusterRoleName"`
	Namespace          string  `json:"namespace"`
	RoleRules          []Rules `json:"roleRules"`
	ClusterRoleRules   []Rules `json:"clusterRoleRules"`
}

// UpdateClusterRoleRequest 定义更新 ClusterRole 的请求结构
type UpdateSaResponse struct {
	handlers.ErrorResponse
}

// UpdateRole 处理更新 Role 的请求
func UpdateSa(w http.ResponseWriter, r *http.Request) {
	var (
		req  UpdateSaRequest
		resp UpdateSaResponse
	)
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求参数失败: %v", err)
		return
	}
	if req.Namespace == "" {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "Namespace 不能为空"
		return
	}

	if req.ServiceAccountName == "" {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "ServiceAccountName 不能为空"
		return
	}

	if req.RoleName == "" && req.ClusterRoleName == "" {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "RoleName 和 ClusterRoleName 不能同时为空"
		return
	}

	// 校验 ClusterRole 相关参数
	if req.ClusterRoleName != "" && len(req.ClusterRoleRules) == 0 {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "指定 ClusterRoleName 时，ClusterRoleRules 不能为空"
		return
	}

	// 校验 Role 相关参数
	if req.RoleName != "" && len(req.RoleRules) == 0 {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "指定 RoleName 时，RoleRules 不能为空"
		return
	}

	// 获取 Kubernetes 客户端
	clientset := k8s.GetClient()

	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), req.Namespace, metav1.GetOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("创建 namespace 失败: %v", err)
		return

	}
	// 检查 ServiceAccount 是否存在
	_, err = clientset.CoreV1().ServiceAccounts(req.Namespace).Get(context.TODO(), req.ServiceAccountName, metav1.GetOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("创建sa 失败: %v", err)
		return
	}

	// 获取现有的 Role
	role, err := clientset.RbacV1().Roles(req.Namespace).Get(context.TODO(), req.RoleName, metav1.GetOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("获取 Role 失败: %v", err)
		return
	}

	var rules []rbacv1.PolicyRule
	for _, rule := range req.RoleRules {
		rules = append(rules, rbacv1.PolicyRule{
			Verbs:     rule.Verbs,
			APIGroups: rule.APIGroups,
			Resources: rule.Resources,
		})
	}
	// 更新 Role 的规则
	role.Rules = rules

	// 应用更新
	_, err = clientset.RbacV1().Roles(req.Namespace).Update(context.TODO(), role, metav1.UpdateOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("更新 Role 失败: %v", err)
		return
	}

	// 获取现有的 ClusterRole
	clusterRole, err := clientset.RbacV1().ClusterRoles().Get(context.TODO(), req.ClusterRoleName, metav1.GetOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("获取 ClusterRole 失败: %v", err)
		return
	}

	var clusterRules []rbacv1.PolicyRule
	for _, rule := range req.ClusterRoleRules {
		clusterRules = append(clusterRules, rbacv1.PolicyRule{
			Verbs:     rule.Verbs,
			APIGroups: rule.APIGroups,
			Resources: rule.Resources,
		})
	}
	// 更新 ClusterRole 的规则
	clusterRole.Rules = clusterRules

	// 应用更新
	_, err = clientset.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("更新 ClusterRole 失败: %v", err)
		return
	}
}
