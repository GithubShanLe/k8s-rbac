package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
)

// CommandRequest 定义命令请求结构体
type CommandRequest struct {
	Command string `json:"command"`
}

// CommandResponse 定义命令响应结构体
type CommandResponse struct {
	Output string `json:"output"`
	Dir    string `json:"dir"`  // 新增字段，用于返回当前目录
	Error  string `json:"error"`
}

// 定义 WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ExecuteCommandHandler 处理执行命令的 WebSocket 请求和普通 HTTP 请求
func ExecuteCommandHandler(w http.ResponseWriter, r *http.Request) {
	var currentDir string // 用于存储当前工作目录
	if r.Header.Get("Upgrade") == "websocket" {
		// 处理 WebSocket 请求
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket 升级失败:", err)
			return
		}
		defer conn.Close()

		for {
			// 读取 WebSocket 消息
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("读取消息失败:", err)
				break
			}

			var req CommandRequest
			err = json.Unmarshal(message, &req)
			if err != nil {
				resp := CommandResponse{
					Error: "请求体解析失败",
				}
				respJSON, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, respJSON)
				continue
			}

			// 解析命令和参数
			parts := strings.Fields(req.Command)
			if len(parts) == 0 {
				resp := CommandResponse{
					Error: "无效的命令",
				}
				respJSON, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, respJSON)
				continue
			}
			cmdName := parts[0]
			cmdArgs := parts[1:]

			if cmdName == "cd" {
				if len(cmdArgs) == 0 {
					resp := CommandResponse{
						Error: "cd 命令缺少目标目录",
					}
					respJSON, _ := json.Marshal(resp)
					conn.WriteMessage(websocket.TextMessage, respJSON)
					continue
				}
				targetDir := cmdArgs[0]
				err := os.Chdir(targetDir)
				if err != nil {
					resp := CommandResponse{
						Error: fmt.Sprintf("无法切换到目录 %s: %v", targetDir, err),
					}
					respJSON, _ := json.Marshal(resp)
					conn.WriteMessage(websocket.TextMessage, respJSON)
					continue
				}
				currentDir = targetDir
				resp := CommandResponse{
					Output: fmt.Sprintf("已切换到目录 %s", targetDir),
					Dir:    currentDir, // 更新当前目录
				}
				respJSON, _ := json.Marshal(resp)
				conn.WriteMessage(websocket.TextMessage, respJSON)
				continue
			}

			// 执行命令
			cmd := exec.Command(cmdName, cmdArgs...)
			if currentDir != "" {
				cmd.Dir = currentDir
			}
			output, err := cmd.CombinedOutput()

			if currentDir == "" {
				currentDir, _ = os.Getwd()
			}

			resp := CommandResponse{
				Output: string(output),
				Dir:    currentDir, // 返回当前目录
			}
			if err != nil {
				resp.Error = err.Error()
			}

			// 发送 JSON 响应
			respJSON, err := json.Marshal(resp)
			if err != nil {
				log.Println("JSON 编码失败:", err)
				break
			}
			err = conn.WriteMessage(websocket.TextMessage, respJSON)
			if err != nil {
				log.Println("发送消息失败:", err)
				break
			}
		}
	} else {
		// 处理普通 HTTP 请求
		if r.Method != http.MethodPost {
			http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
			return
		}

		var req CommandRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "请求体解析失败", http.StatusBadRequest)
			return
		}

		// 解析命令和参数
		parts := strings.Fields(req.Command)
		if len(parts) == 0 {
			http.Error(w, "无效的命令", http.StatusBadRequest)
			return
		}
		cmdName := parts[0]
		cmdArgs := parts[1:]

		if cmdName == "cd" {
			if len(cmdArgs) == 0 {
				http.Error(w, "cd 命令缺少目标目录", http.StatusBadRequest)
				return
			}
			targetDir := cmdArgs[0]
			err := os.Chdir(targetDir)
			if err != nil {
				http.Error(w, fmt.Sprintf("无法切换到目录 %s: %v", targetDir, err), http.StatusBadRequest)
				return
			}
			currentDir = targetDir
			resp := CommandResponse{
				Output: fmt.Sprintf("已切换到目录 %s", targetDir),
				Dir:    currentDir, // 更新当前目录
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// 执行命令
		cmd := exec.Command(cmdName, cmdArgs...)
		if currentDir != "" {
			cmd.Dir = currentDir
		}
		output, err := cmd.CombinedOutput()

		if currentDir == "" {
			currentDir, _ = os.Getwd()
		}

		resp := CommandResponse{
			Output: string(output),
			Dir:    currentDir, // 返回当前目录
		}
		if err != nil {
			resp.Error = err.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// StartServer 启动 WebSocket 和 HTTP 服务器
func StartServer(port int) {
	http.HandleFunc("/execute/shell", ExecuteCommandHandler)
	log.Printf("WebSocket 和 HTTP API 服务器启动，监听端口: %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
