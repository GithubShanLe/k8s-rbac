package nodepool

import (
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func GetNodeStatus(node corev1.Node) string {
	statuses := make(sort.StringSlice, 10)
	status(node.Status.Conditions, node.Spec.Unschedulable, statuses)
	return join(statuses, ", ")
}

func status(conds []corev1.NodeCondition, exempt bool, res []string) {
	var index int
	conditions := make(map[corev1.NodeConditionType]*corev1.NodeCondition, len(conds))
	for n := range conds {
		cond := conds[n]
		conditions[cond.Type] = &cond
	}

	validConditions := []corev1.NodeConditionType{corev1.NodeReady}
	for _, validCondition := range validConditions {
		condition, ok := conditions[validCondition]
		if !ok {
			continue
		}
		neg := ""
		if condition.Status != corev1.ConditionTrue {
			neg = "Not"
		}
		res[index] = neg + string(condition.Type)
		index++
	}
	if len(res) == 0 {
		res[index] = "Unknown"
		index++
	}
	if exempt {
		res[index] = "SchedulingDisabled"
	}
}

func empty(s []string) bool {
	for _, v := range s {
		if len(v) != 0 {
			return false
		}
	}
	return true
}

func join(a []string, sep string) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return a[0]
	}

	b := make([]string, 0, len(a))
	for _, s := range a {
		if s != "" {
			b = append(b, s)
		}
	}
	if len(b) == 0 {
		return ""
	}

	n := len(sep) * (len(b) - 1)
	for i := 0; i < len(b); i++ {
		n += len(a[i])
	}

	var buff strings.Builder
	buff.Grow(n)
	buff.WriteString(b[0])
	for _, s := range b[1:] {
		buff.WriteString(sep)
		buff.WriteString(s)
	}

	return buff.String()
}
