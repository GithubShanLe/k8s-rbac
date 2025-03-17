package handlers

import (
	"encoding/json"
	"fmt"
	"k8s-rbac-backend/k8s"
	"net/http"

	"k8s.io/client-go/discovery"
)

// 定义资源类型结构体
type APIResource struct {
	Name       string   `json:"name"`
	Namespaced bool     `json:"namespaced"`
	Kind       string   `json:"kind"`
	APIGroup   []string `json:"apiGroup"`
}

type GetResourcesResponse struct {
	ErrorResponse
	Resources []APIResource `json:"resources"`
	ApiGroups []string      `json:"apiGroups"`
}

// 获取所有支持的资源类型
func GetResources(w http.ResponseWriter, r *http.Request) {
	var resp GetResourcesResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()
	// 获取 Discovery 客户端
	discoveryClient := k8s.GetDiscoveryClient()

	// 获取所有 API 资源
	// apiResourceLists, err := discoveryClient.ServerPreferredResources()
	// if err != nil {
	// 	resp.ErrorCode = "400"
	// 	resp.ErrorMessage = fmt.Sprintf("获取资源列表失败: %v", err)
	// 	return
	// }

	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		if !discovery.IsGroupDiscoveryFailedError(err) {
			resp.ErrorCode = "400"
			resp.ErrorMessage = fmt.Sprintf("获取资源列表失败: %v", err)
			return
		}
	}
	var (
		apiResources = make(map[string]APIResource)
		apiGroups    = make(map[string]string)
	)
	// 提取资源信息
	for _, apiResourceList := range apiResourceLists {
		groupVersion := apiResourceList.GroupVersion
		for _, apiResource := range apiResourceList.APIResources {
			// 解析 API 组
			apiGroup := ""
			if groupVersion != "v1" {
				apiGroup = groupVersion
			}
			if _, ok := apiGroups[apiGroup]; !ok {
				apiGroups[apiGroup] = apiGroup
			}
			if _, ok := apiResources[apiResource.Name]; !ok {
				apiResources[apiResource.Name] = APIResource{
					Name:       apiResource.Name,
					Namespaced: apiResource.Namespaced,
					Kind:       apiResource.Kind,
					APIGroup:   []string{apiGroup},
				}
			} else {
				a := apiResources[apiResource.Name]
				a.APIGroup = append(a.APIGroup, apiGroup)
				apiResources[apiResource.Name] = a
			}
		}
	}
	for idx := range apiResources {
		resp.Resources = append(resp.Resources, apiResources[idx])
	}
	resp.Resources = append(resp.Resources, APIResource{
		Name:       "*",
		Namespaced: true,
	})
	for idx := range apiGroups {
		resp.ApiGroups = append(resp.ApiGroups, apiGroups[idx])
	}
	return
}
