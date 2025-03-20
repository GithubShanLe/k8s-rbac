package nodepool

var nodePoolNameMap = map[string]bool{
	"alibabacloud.com/nodepool-id":  true,
	"cloud.google.com/gke-nodepool": true,
	"cce.cloud.com/cce-nodepool":    true,
	"eks.amazonaws.com/nodegroup":   true,
	"nodepool":                      true,
}
