package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/k8s"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccountInfo 结构体用于存储 ServiceAccount 及其关联的 RBAC 资源信息
type ServiceAccountInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// Secrets         []corev1.Secret             `json:"secrets"`
	// Roles           []rbacv1.Role               `json:"roles"`
	// ClusterRoles    []rbacv1.ClusterRole        `json:"clusterRoles"`
	// RoleBindings    []rbacv1.RoleBinding        `json:"roleBindings"`
	// ClusterBindings []rbacv1.ClusterRoleBinding `json:"clusterBindings"`
	Roles        []Role        `json:"roles"`
	ClusterRoles []ClusterRole `json:"clusterRoles"`
}

type Role struct {
	RoleName        string     `json:"roleName"`
	NameSpace       string     `json:"nameSpace"`
	RoleAge         string     `json:"roleAge"`
	RoleBindingName string     `json:"roleBindingName"`
	RoleBindingAge  string     `json:"roleBindingAge"`
	Rules           []RuleInfo `json:"rules"`
}

type ClusterRole struct {
	ClusterRoleName        string     `json:"clusterRoleName"`
	NameSpace              string     `json:"nameSpace"`
	ClusterRoleAge         string     `json:"clusterRoleAge"`
	ClusterRoleBindingName string     `json:"clusterRoleBindingName"`
	ClusterRoleBindingAge  string     `json:"clusterRoleBindingAge"`
	Rules                  []RuleInfo `json:"rules"`
}

type RuleInfo struct {
	Verbs    []string `json:"verbs"`
	ApiGroup []string `json:"apiGroup"`
	Resource string   `json:"resource"`
}

type ListServiceAccountsRequest struct {
	Namespace string `json:"namespace"` //不填返回所有的namespace 的sa
}

type ListServiceAccountsResponse struct {
	ErrorResponse
	ServiceAccounts []ServiceAccount `json:"serviceAccounts"`
}

type ServiceAccount struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	SecretName []string `json:"secretName"`
	Age        string   `json:"age"`
}

// ListServiceAccounts 获取所有 ServiceAccount 的基本信息
func ListServiceAccounts(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var resp ListServiceAccountsResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()
	var req ListServiceAccountsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	// 获取 Kubernetes 客户端
	clientset := k8s.GetClient()
	// 获取所有命名空间
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), req.Namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.ErrorCode = "404"
			resp.ErrorMessage = fmt.Sprintf("命名空间 %s 不存在", req.Namespace)
		} else {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("获取命名空间失败: %v", err)
		}
		return
	}
	// 获取当前命名空间下的所有 ServiceAccount
	sas, err := clientset.CoreV1().ServiceAccounts(ns.Name).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}
	// 遍历 ServiceAccount
	for _, sa := range sas.Items {
		resp.ServiceAccounts = append(resp.ServiceAccounts, ServiceAccount{
			Name:      sa.Name,
			Namespace: sa.Namespace,
			SecretName: func() []string {
				var secrets []string
				for _, item := range sa.Secrets {
					secrets = append(secrets, item.Name)
				}
				return secrets
			}(),
			// 计算创建时间到现在的小时数
			// 计算创建时间到现在的时间差并格式化为天时分秒
			Age: func() string {
				duration := time.Since(sa.CreationTimestamp.Time)
				days := int(duration.Hours()) / 24
				hours := int(duration.Hours()) % 24
				minutes := int(duration.Minutes()) % 60
				seconds := int(duration.Seconds()) % 60
				return fmt.Sprintf("%dd%dh%dm%ds", days, hours, minutes, seconds)
			}(),
		})
	}
}

// GetServiceAccountDetails 获取指定 ServiceAccount 的详细信息

type GetServiceAccountDetailsResponse struct {
	ErrorResponse
	ServiceAccountInfo ServiceAccountInfo `json:"serviceAccountInfo"`
}

type GetServiceAccountDetailsRequest struct {
	ServiceAccountName string `json:"serviceAccountName"`
	Namespace          string `json:"namespace"`
}

// 获取sa详情
func GetServiceAccountDetails(w http.ResponseWriter, r *http.Request) {
	var resp GetServiceAccountDetailsResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()
	var req GetServiceAccountDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}
	if req.ServiceAccountName == "" {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("ServiceAccountName 不能为空")
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	// 获取 Kubernetes 客户端
	clientset := k8s.GetClient()
	// 获取 ServiceAccount
	sa, err := clientset.CoreV1().ServiceAccounts(req.Namespace).Get(context.TODO(), req.ServiceAccountName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.ErrorCode = "404"
			resp.ErrorMessage = fmt.Sprintf("service account %s not exit", req.ServiceAccountName)
		} else {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("获取service account failed: %v", err)
		}
		return
	}

	resp.ServiceAccountInfo.Name = sa.Name
	resp.ServiceAccountInfo.Namespace = sa.Namespace

	// // 获取 Secret
	// for _, secretRef := range sa.Secrets {
	// 	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretRef.Name, metav1.GetOptions{})
	// 	if err == nil {
	// 		saInfo.Secrets = append(saInfo.Secrets, *secret)
	// 	}
	// }
	var roles = make(map[string]Role)
	var clusterRoles = make(map[string]ClusterRole)
	// 获取 RoleBinding
	roleBindings, err := clientset.RbacV1().RoleBindings(req.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, rb := range roleBindings.Items {
			for _, subject := range rb.Subjects {
				if subject.Kind == "ServiceAccount" && subject.Name == req.ServiceAccountName && subject.Namespace == req.Namespace {
					// resp.ServiceAccountInfo.RoleBindings = append(resp.ServiceAccountInfo.RoleBindings, rb)

					// 获取关联的 Role
					role, err := clientset.RbacV1().Roles(req.Namespace).Get(context.TODO(), rb.RoleRef.Name, metav1.GetOptions{})
					if err != nil {
						resp.ErrorCode = "400"
						resp.ErrorMessage = fmt.Sprintf("获取service account role failed: %v", err)
						return
					}
					if _, ok := roles[role.Name+role.Namespace]; !ok {
						ro := Role{
							RoleName:        role.Name,
							RoleAge:         role.CreationTimestamp.String(),
							NameSpace:       subject.Namespace,
							RoleBindingName: rb.Name,
							RoleBindingAge:  rb.CreationTimestamp.String(),
							Rules:           []RuleInfo{},
						}
						for _, rule := range role.Rules {
							for _, ruleSource := range rule.Resources {
								ro.Rules = append(ro.Rules, RuleInfo{
									Verbs:    rule.Verbs,
									ApiGroup: rule.APIGroups,
									Resource: ruleSource,
								})
							}
						}
						roles[role.Name+role.Namespace] = ro
					} else {
						for _, rule := range role.Rules {
							for _, ruleSource := range rule.Resources {
								if len(rule.Verbs) > 0 {
									// 获取当前角色
									roleTemp := roles[role.Name+role.Namespace]
									// 添加新的规则
									roleTemp.Rules = append(roleTemp.Rules, RuleInfo{
										Verbs:    rule.Verbs,
										ApiGroup: rule.APIGroups,
										Resource: ruleSource,
									})
									// 更新 map 中的角色
									roles[role.Name+role.Namespace] = roleTemp
								}
							}
						}
					}

				}
			}
		}
	}

	// 获取 ClusterRoleBinding
	clusterRoleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, crb := range clusterRoleBindings.Items {
			for _, subject := range crb.Subjects {
				if subject.Kind == "ServiceAccount" && subject.Name == req.ServiceAccountName && subject.Namespace == req.Namespace {
					// resp.ServiceAccountInfo.ClusterBindings = append(resp.ServiceAccountInfo.ClusterBindings, crb)

					// // 获取关联的 ClusterRole
					clusterRole, err := clientset.RbacV1().ClusterRoles().Get(context.TODO(), crb.RoleRef.Name, metav1.GetOptions{})
					if err != nil {
						resp.ErrorCode = "400"
						resp.ErrorMessage = fmt.Sprintf("获取service account clustr role failed: %v", err)
						return
					}
					if _, ok := clusterRoles[clusterRole.Name+clusterRole.Namespace]; !ok {
						clusterRo := ClusterRole{
							ClusterRoleName:        clusterRole.Name,
							ClusterRoleAge:         clusterRole.CreationTimestamp.String(),
							NameSpace:              subject.Namespace,
							ClusterRoleBindingName: crb.Name,
							ClusterRoleBindingAge:  crb.CreationTimestamp.String(),
							Rules:                  []RuleInfo{},
						}
						for _, rule := range clusterRole.Rules {
							for _, ruleSource := range rule.Resources {
								clusterRo.Rules = append(clusterRo.Rules, RuleInfo{
									Verbs:    rule.Verbs,
									ApiGroup: rule.APIGroups,
									Resource: ruleSource,
								})
							}
						}
						clusterRoles[clusterRole.Name+clusterRole.Namespace] = clusterRo
					} else {
						for _, rule := range clusterRole.Rules {
							for _, ruleSource := range rule.Resources {
								if len(rule.Verbs) > 0 {
									// 获取当前角色
									clusterRoleTemp := clusterRoles[clusterRole.Name+clusterRole.Namespace]
									// 添加新的规则
									clusterRoleTemp.Rules = append(clusterRoleTemp.Rules, RuleInfo{
										Verbs:    rule.Verbs,
										ApiGroup: rule.APIGroups,
										Resource: ruleSource,
									})
									// 更新 map 中的角色
									clusterRoles[clusterRole.Name+clusterRole.Namespace] = clusterRoleTemp
								}
							}
						}
					}
				}
			}
		}
	}
	for _, v := range roles {
		resp.ServiceAccountInfo.Roles = append(resp.ServiceAccountInfo.Roles, v)
	}
	for _, v := range clusterRoles {
		resp.ServiceAccountInfo.ClusterRoles = append(resp.ServiceAccountInfo.ClusterRoles, v)
	}
	return
}
