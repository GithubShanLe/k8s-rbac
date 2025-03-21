/*
 * @Author: le.shan le.shan@transsion.com
 * @Date: 2025-03-20 17:07:52
 * @LastEditors: le.shan le.shan@transsion.com
 * @LastEditTime: 2025-03-21 18:05:43
 * @FilePath: /k8s-rbac/k8s-rbac-backend/k8s/client.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

var clientset *kubernetes.Clientset
var discoveryClient *discovery.DiscoveryClient
var metriclient *versioned.Clientset

// 初始化 Kubernetes 客户端
func init() {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		panic("无法找到 kubeconfig 文件")
	}
	kubeconfig = "/Users/shanyue/kubecm_config/sg_prod"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Sprintf("无法加载 kubeconfig: %v", err))
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("无法创建 Kubernetes 客户端: %v", err))
	}

	discoveryClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("无法创建 Discovery 客户端: %v", err))
	}

	metriclient = versioned.NewForConfigOrDie(config)
}

// 获取 Kubernetes 客户端
func GetClient() *kubernetes.Clientset {
	return clientset
}

// 获取 Discovery 客户端
func GetDiscoveryClient() *discovery.DiscoveryClient {
	return discoveryClient
}

func GetMetricClient() *versioned.Clientset {
	return metriclient
}
