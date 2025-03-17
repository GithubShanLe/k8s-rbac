package handlers

import (
	"encoding/json"
	"net/http"
)

// 获取所有支持的 verbs
func GetVerbs(w http.ResponseWriter, r *http.Request) {
	verbs := []string{
		"get",
		"list",
		"watch",
		"create",
		"update",
		"patch",
		"delete",
		"deletecollection",
	}

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verbs)
}
