// 定义 WebSocket 连接地址，根据 vue.config.js 中的代理配置修改
const WEBSOCKET_URL_TERMINAL = '/execute/shell';

// 创建一个 WebSocket 实例
let socket;
let isWebSocketOpen = false;
let resolveCallback;

// 初始化 WebSocket 连接
function initWebSocket() {
  socket = new WebSocket(WEBSOCKET_URL_TERMINAL);

  // 监听 WebSocket 连接打开事件
  socket.onopen = () => {
    console.log('WebSocket 连接已打开');
    isWebSocketOpen = true;
  };

  // 监听 WebSocket 接收到消息事件
  socket.onmessage = (event) => {
    console.log('收到消息:', event.data);
    if (resolveCallback) {
      resolveCallback(event.data);
      resolveCallback = null;
    }
  };

  // 监听 WebSocket 连接关闭事件
  socket.onclose = () => {
    console.log('WebSocket 连接已关闭');
    isWebSocketOpen = false;
    // 可以在这里处理重连逻辑
  };

  // 监听 WebSocket 连接错误事件
  socket.onerror = (error) => {
    console.error('WebSocket 连接发生错误:', error);
    isWebSocketOpen = false;
  };
}

// 发送数据到后端
export function withTerminal(data) {
  if (isWebSocketOpen) {
    socket.send(JSON.stringify(data));
    return new Promise((resolve) => {
      resolveCallback = resolve;
    });
  } else {
    console.error('WebSocket 连接未打开');
    return Promise.reject(new Error('WebSocket 连接未打开'));
  }
}

// 初始化 WebSocket 连接
initWebSocket();