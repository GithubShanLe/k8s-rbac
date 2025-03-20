package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *kubernetes.Clientset
var discoveryClient *discovery.DiscoveryClient

// 初始化 Kubernetes 客户端
func init() {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		panic("无法找到 kubeconfig 文件")
	}
	kubeconfig = "/Users/shanyue/kubecm_config/ops-prod-alisg"
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
}

// 获取 Kubernetes 客户端
func GetClient() *kubernetes.Clientset {
	return clientset
}

// 获取 Discovery 客户端
func GetDiscoveryClient() *discovery.DiscoveryClient {
	return discoveryClient
}
