/*
 * @Author: le.shan le.shan@transsion.com
 * @Date: 2025-03-21 15:41:20
 * @LastEditors: le.shan le.shan@transsion.com
 * @LastEditTime: 2025-03-21 16:11:13
 * @FilePath: /k8s-rbac/k8s-rbac-backend/handlers/workload/CronJob.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
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

type ListCronJobRequest struct {
	CronJobName string `json:"CronJobName"`
	NameSpace   string `json:"namespace"`
}

type ListCronJobResponse struct {
	handlers.ErrorResponse
	CronJobs []CronJob `json:"cronJobs"`
}

type CronJob struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Labels     map[string]string `json:"labels"`
	Images     []string          `json:"images"`
	Pods       string            `json:"pods"`
	CreateTime string            `json:"createTime"`
}

func ListCronJob(w http.ResponseWriter, r *http.Request) {
	clientset := k8s.GetClient()
	var resp ListCronJobResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()

	// 解析请求参数
	var req ListCronJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	listOptions := metav1.ListOptions{}
	if req.CronJobName != "" {
		listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", req.CronJobName)
	}
	svcs, err := clientset.BatchV1().CronJobs(req.NameSpace).List(context.Background(), listOptions)
	if err != nil {
		resp.ErrorCode = "500"
		resp.ErrorMessage = fmt.Sprintf("获取svc列表失败: %v", err)
		return
	}
	for _, svc := range svcs.Items {
		resp.CronJobs = append(resp.CronJobs, CronJob{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Labels:    svc.Labels,
			Images: func() []string {
				var images []string
				for _, container := range svc.Spec.JobTemplate.Spec.Template.Spec.Containers {
					images = append(images, container.Image)
				}
				return images
			}(),
			CreateTime: svc.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}
	return
}
