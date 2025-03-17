package models

import rbacv1 "k8s.io/api/rbac/v1"

// 定义请求参数结构体
type CreateSARequest struct {
	ServiceAccountName string              `json:"serviceAccountName"`
	Namespace          string              `json:"namespace"`
	RoleRules          []rbacv1.PolicyRule `json:"roleRules"`
	ClusterRoleRules   []rbacv1.PolicyRule `json:"clusterRoleRules"`
}
