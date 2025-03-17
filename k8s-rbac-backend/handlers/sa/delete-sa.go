package sa

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/handlers"
	"k8s-rbac-backend/k8s"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeleteSARequest struct {
	ServiceAccountName     string `json:"serviceAccountName"`
	RoleName               string `json:"roleName"`
	RoleBindingName        string `json:"roleBindingName"`
	ClusterRoleName        string `json:"clusterRoleName"`
	ClusterRoleBindingName string `json:"clusterRoleBindingName"`
	Namespace              string `json:"namespace"`
}

type DeleteSAResponse struct {
	handlers.ErrorResponse
}

func DeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	var (
		req  DeleteSARequest
		resp DeleteSAResponse
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
	clientset := k8s.GetClient()
	_, err := clientset.CoreV1().Namespaces().Get(context.TODO(), req.Namespace, metav1.GetOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("获取 namespace 失败: %v", err)
		return
	}
	// 检查 ServiceAccount 是否存在,不存在不报错
	_, err = clientset.CoreV1().ServiceAccounts(req.Namespace).Get(context.TODO(), req.ServiceAccountName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("获取失败: %v", err)
			return
		}
	}
	//获取rolebinding
	rb, err := clientset.RbacV1().RoleBindings(req.Namespace).Get(context.TODO(), req.RoleBindingName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("获取失败: %v", err)
			return
		}
	} else {
		// 通过rolebinging删除role
		for _, item := range rb.Subjects {
			if item.Kind == "ServiceAccount" && item.Name == req.ServiceAccountName && item.Namespace == req.Namespace && rb.RoleRef.Kind == "Role" && rb.RoleRef.Name == req.RoleName {
				err = clientset.RbacV1().RoleBindings(req.Namespace).Delete(context.TODO(), req.RoleBindingName, metav1.DeleteOptions{})
				if err != nil {
					resp.ErrorCode = "400"
					resp.ErrorMessage = fmt.Sprintf("删除 RoleBinding 失败: %v", err)
					return
				}
				var flag bool
				//获取role
				_, err = clientset.RbacV1().Roles(req.Namespace).Get(context.TODO(), req.RoleName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						flag = true
					} else {
						resp.ErrorCode = "400"
						resp.ErrorMessage = fmt.Sprintf("获取失败: %v", err)
						return
					}
				}
				if !flag {
					err = clientset.RbacV1().Roles(req.Namespace).Delete(context.TODO(), req.RoleName, metav1.DeleteOptions{})
					if err != nil {
						resp.ErrorCode = "400"
						resp.ErrorMessage = fmt.Sprintf("删除 Role 失败: %v", err)
						return
					}
				}

			}
		}
	}

	//获取clusterrolebinding
	crb, err := clientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), req.ClusterRoleBindingName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("获取失败: %v", err)
			return
		}
	} else {
		// 通过clusterrolebinging删除clusterrole
		for _, item := range crb.Subjects {
			if item.Kind == "ServiceAccount" && item.Name == req.ServiceAccountName && item.Namespace == req.Namespace && crb.RoleRef.Kind == "ClusterRole" && crb.RoleRef.Name == req.ClusterRoleName {
				err = clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), req.ClusterRoleBindingName, metav1.DeleteOptions{})
				if err != nil {
					resp.ErrorCode = "400"
					resp.ErrorMessage = fmt.Sprintf("删除 ClusterRoleBinding 失败: %v", err)
					return
				}
				var flag bool
				//获取clusterrole
				_, err = clientset.RbacV1().ClusterRoles().Get(context.TODO(), req.ClusterRoleName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						flag = true
					} else {
						resp.ErrorCode = "400"
						resp.ErrorMessage = fmt.Sprintf("获取失败: %v", err)
						return
					}
				}
				if !flag {
					err = clientset.RbacV1().ClusterRoles().Delete(context.TODO(), req.ClusterRoleName, metav1.DeleteOptions{})
					if err != nil {
						resp.ErrorCode = "400"
						resp.ErrorMessage = fmt.Sprintf("删除 ClusterRole 失败: %v", err)
						return
					}
				}
			}
		}
	}
}
