package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleAllNamespace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只处理包含请求体的请求
		if r.Body != nil {
			// 读取请求体
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "读取请求体失败", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			// 解析为 map 以便处理任意 JSON 结构
			var requestData map[string]interface{}
			if err := json.Unmarshal(body, &requestData); err == nil {
				// 递归处理所有字段
				modified := processAllPlus(requestData)

				// 重新编码处理后的数据
				newBody, err := json.Marshal(modified)
				if err != nil {
					http.Error(w, "处理请求数据失败", http.StatusInternalServerError)
					return
				}

				// 替换请求体
				r.Body = io.NopCloser(bytes.NewBuffer(newBody))
				r.ContentLength = int64(len(newBody))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func processAllPlus(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// 处理对象
		for key, value := range v {
			if str, ok := value.(string); ok {
				lowKey := strings.ToLower(key)
				if (lowKey == "ns" || lowKey == "namespace") && str == "all+" {
					v[key] = ""
				}
			} else {
				v[key] = processAllPlus(value)
			}
		}
		return v
	case []interface{}:
		// 处理数组
		for i, value := range v {
			v[i] = processAllPlus(value)
		}
		return v
	default:
		return v
	}
}
