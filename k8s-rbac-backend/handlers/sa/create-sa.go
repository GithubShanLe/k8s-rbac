package sa

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 定义请求参数结构体
type CreateSARequest struct {
	ServiceAccountName string  `json:"serviceAccountName"`
	RoleName           string  `json:"roleName"`
	ClusterRoleName    string  `json:"clusterRoleName"`
	Namespace          string  `json:"namespace"`
	RoleRules          []Rules `json:"roleRules"`
	ClusterRoleRules   []Rules `json:"clusterRoleRules"`
}
type CreateSAResponse struct {
	handlers.ErrorResponse
}

type Rules struct {
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// 创建 ServiceAccount 及相关 RBAC 资源的接口
func CreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var (
		req  CreateSARequest
		resp CreateSAResponse
	)
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求参数失败: %v", err)
		return
	}
	// 校验请求参数
	if req.ServiceAccountName == "" {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "ServiceAccountName 不能为空"
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

	// 至少需要指定 Role 或 ClusterRole 中的一个
	if req.RoleName == "" && req.ClusterRoleName == "" {
		resp.ErrorCode = "400"
		resp.ErrorMessage = "RoleName 和 ClusterRoleName 不能同时为空"
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	// 获取 Kubernetes 客户端
	// 检查 namespace 是否存在，不存在则创建
	clientset := k8s.GetClient()
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), req.Namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// 创建 namespace
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: req.Namespace,
				},
			}
			_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
			if err != nil {
				resp.ErrorCode = "400"
				resp.ErrorMessage = fmt.Sprintf("创建 namespace 失败: %v", err)
				return
			}
		} else {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("创建 namespace 失败: %v", err)
			return
		}
	}

	// 创建 ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.ServiceAccountName,
			Namespace: req.Namespace,
		},
	}
	// 检查 ServiceAccount 是否存在
	_, err = clientset.CoreV1().ServiceAccounts(req.Namespace).Get(context.TODO(), req.ServiceAccountName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = clientset.CoreV1().ServiceAccounts(req.Namespace).Create(context.TODO(), sa, metav1.CreateOptions{})
			if err != nil {
				resp.ErrorCode = "400"
				resp.ErrorMessage = fmt.Sprintf("创建sa 失败: %v", err)
				return
			}
		} else {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("创建sa 失败: %v", err)
			return
		}
	}

	// 创建 Role
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.RoleName,
			Namespace: req.Namespace,
		},
		Rules: func() []rbacv1.PolicyRule {
			rules := make([]rbacv1.PolicyRule, 0, len(req.RoleRules))
			for _, rule := range req.RoleRules {
				rules = append(rules, rbacv1.PolicyRule{
					APIGroups: rule.APIGroups,
					Resources: rule.Resources,
					Verbs:     rule.Verbs,
				})
			}
			return rules
		}(),
	}

	_, err = clientset.RbacV1().Roles(req.Namespace).Create(context.TODO(), role, metav1.CreateOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("创建 Role 失败: %v", err)})
		return
	}

	// 创建 RoleBinding
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-rb", req.RoleName),
			Namespace: req.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      req.ServiceAccountName,
				Namespace: req.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     role.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = clientset.RbacV1().RoleBindings(req.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("创建 RoleBinding 失败: %v", err)})
		return
	}

	// 创建 ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.ClusterRoleName,
		},
		Rules: func() []rbacv1.PolicyRule {
			rules := make([]rbacv1.PolicyRule, 0, len(req.ClusterRoleRules))
			for _, rule := range req.ClusterRoleRules {
				rules = append(rules, rbacv1.PolicyRule{
					APIGroups: rule.APIGroups,
					Resources: rule.Resources,
					Verbs:     rule.Verbs,
				})
			}
			return rules
		}(),
	}

	_, err = clientset.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("创建 ClusterRole 失败: %v", err)})
		return
	}

	// 创建 ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-crb", req.ClusterRoleName),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      req.ServiceAccountName,
				Namespace: req.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("创建 ClusterRoleBinding 失败: %v", err)})
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "ServiceAccount %s 及相关 RBAC 资源创建成功\n", req.ServiceAccountName)
}
