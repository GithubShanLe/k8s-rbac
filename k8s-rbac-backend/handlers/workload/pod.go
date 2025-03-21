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

type ListPodRequest struct {
	PodName   string `json:"podName"`
	NameSpace string `json:"namespace"`
}

type ListPodResponse struct {
	handlers.ErrorResponse
	Pods []Pod `json:"pods"`
}

type Pod struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Labels     map[string]string `json:"labels"`
	Images     []string          `json:"images"`
	Cpu        string            `json:"cpu"`
	Mem        string            `json:"mem"`
	Pods       string            `json:"pods"`
	Restart    int32             `json:"restart"`
	CreateTime string            `json:"createTime"`
}

func ListPod(w http.ResponseWriter, r *http.Request) {
	clientset := k8s.GetClient()
	var resp ListPodResponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()

	// 解析请求参数
	var req ListPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}

	listOptions := metav1.ListOptions{}
	if req.PodName != "" {
		listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", req.PodName)
	}
	pods, err := clientset.CoreV1().Pods(req.NameSpace).List(context.Background(), listOptions)
	if err != nil {
		resp.ErrorCode = "500"
		resp.ErrorMessage = fmt.Sprintf("获取svc列表失败: %v", err)
		return
	}
	for _, pod := range pods.Items {
		resp.Pods = append(resp.Pods, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
			Images: func() []string {
				var images []string
				for _, container := range pod.Spec.Containers {
					images = append(images, container.Image)
				}
				return images
			}(),
			CreateTime: pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Cpu: func() string {
				var (
					requestCpu int64
					limitCpu   int64
				)

				for _, container := range pod.Spec.Containers {
					requestCpu += container.Resources.Requests.Cpu().MilliValue()
					limitCpu += container.Resources.Limits.Cpu().MilliValue()
				}
				return fmt.Sprintf("%d/%d", requestCpu, limitCpu)
			}(),
			Mem: func() string {
				var (
					limitMem   int64
					requestMem int64
				)
				for _, container := range pod.Spec.Containers {
					limitMem += container.Resources.Limits.Memory().Value() / 1024 / 1024
					requestMem += container.Resources.Requests.Memory().Value() / 1024 / 1024
				}
				return fmt.Sprintf("%d/%d", requestMem, limitMem)
			}(),
			Restart: func() int32 {
				var restart int32
				for _, container := range pod.Status.ContainerStatuses {
					restart += container.RestartCount
				}
				return restart
			}(),
		})
	}
	return
}

type GetPodMetricRequest struct {
	PodName   string `json:"PodName"`
	NameSpace string `json:"namespace"`
}

type GetPodMetricsponse struct {
	handlers.ErrorResponse
	Metric Metric `json:"metric"`
}
type Metric struct {
	CpuLimitRate      string `json:"cpuLimitRate"`
	CpuRequestRate    string `json:"cpuRequestRate"`
	MemoryLimitRate   string `json:"memoryLimitRate"`
	MemoryReuqestRate string `json:"memoryReuqestRate"`
}

func GetPodMetric(w http.ResponseWriter, r *http.Request) {

	var resp GetPodMetricsponse
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}()

	// 解析请求参数
	var req GetPodMetricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("解析请求失败: %v", err)
		return
	}
	clientset := k8s.GetClient()
	pod, err := clientset.CoreV1().Pods(req.NameSpace).Get(context.TODO(), req.PodName, metav1.GetOptions{})
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("get pod err: %s", err.Error())
		return
	}
	// 获取cpu的请求值
	var cpuRequest, cpuLimit, memoryRequest, memoryLimit int64
	for _, containers := range pod.Spec.Containers {
		cpuRequest += containers.Resources.Requests.Cpu().MilliValue()
		cpuLimit += containers.Resources.Limits.Cpu().MilliValue()
		memoryRequest += containers.Resources.Requests.Memory().MilliValue()
		memoryLimit += containers.Resources.Limits.Memory().MilliValue()
	}

	getOptions := metav1.GetOptions{}
	metricClientset := k8s.GetMetricClient()
	result, err := metricClientset.MetricsV1beta1().PodMetricses(req.NameSpace).Get(context.TODO(), req.PodName, getOptions)
	if err != nil {
		resp.ErrorCode = "400"
		resp.ErrorMessage = fmt.Sprintf("get podMetricses err: %s", err.Error())
		return
	}
	var (
		memoryLimitRate, memoryReuqestRate string
		cpuLimitRate, cpuRequestRate       string
		cpu, mem                           int64
	)

	for _, container := range result.Containers {
		cpu += container.Usage.Cpu().MilliValue()
		mem += container.Usage.Memory().MilliValue()
	}
	// 处理分母为0的情况，返回N/A
	if cpuLimit == 0 {
		cpuLimitRate = "N/A"
	} else {
		cpuLimitRate = fmt.Sprintf("%.1f", (float64(cpu)/float64(cpuLimit))*100)
	}
	fmt.Println(cpuLimit, cpuRequest, memoryRequest, memoryLimit, cpu, mem)
	if cpuRequest == 0 {
		cpuRequestRate = "N/A"
	} else {
		cpuRequestRate = fmt.Sprintf("%.1f", (float64(cpu)/float64(cpuRequest))*100)
	}

	if memoryLimit == 0 {
		memoryLimitRate = "N/A"
	} else {
		memoryLimitRate = fmt.Sprintf("%.1f", (float64(mem)/float64(memoryLimit))*100)
	}

	if memoryRequest == 0 {
		memoryReuqestRate = "N/A"
	} else {
		memoryReuqestRate = fmt.Sprintf("%.1f", (float64(mem)/float64(memoryRequest))*100)
	}
	resp.Metric = Metric{
		CpuLimitRate:      cpuLimitRate,
		CpuRequestRate:    cpuRequestRate,
		MemoryLimitRate:   memoryLimitRate,
		MemoryReuqestRate: memoryReuqestRate,
	}
	return
}
