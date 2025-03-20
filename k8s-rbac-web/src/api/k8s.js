// 引入 axios 库
import axios from 'axios';
// 定义 API 基础路径
const API_BASE_URL = '/api'; // 可根据环境动态调整
// 将 axios 配置为基础路径
const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 50000, // 设置超时时间
});

export function getNamespaces() {
  return request({
    url: '/ns', // 相对路径，已包含在 baseURL 中
    method: 'get'
  })
  .then(response => {
    // 校验后端返回的数据结构是否符合预期
    if ( response && response.status === 200 && Array.isArray(response.data.namespaces)) {
      return response.data.namespaces; // 返回有效的数据部分
    } else {
      throw new Error('Invalid response format from server');
    }
  })
  .catch(error => {
    // 捕获并处理异常，避免未捕获的错误
    console.error('Error fetching namespaces:', error);
    return []; // 返回空数组作为默认值，防止调用方收到 undefined
  });
}
export function getServiceAccounts(ns) {
    return request({
      url: '/listSa', // 相对路径，已包含在 baseURL 中
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        "namespace":ns
      },
    })
    .then(response => {
        console.log(response)
      // 校验后端返回的数据结构是否符合预期
      return response
    })
    .catch(error => {
      // 捕获并处理异常，避免未捕获的错误
      console.error('Error fetching sa:', error);
      return []; // 返回空数组作为默认值，防止调用方收到 undefined
    });
  }

export function getSaRoles(namespace, name) {
  return request({
    url: '/sa-details',
    method: 'post',
    data: {
        serviceAccountName:name,
        namespace:namespace
    }
  })
  .then(response => {
    console.log(response)
  // 校验后端返回的数据结构是否符合预期
  return response
})
.catch(error => {
  // 捕获并处理异常，避免未捕获的错误
  console.error('Error fetching sa:', error);
  return []; // 返回空数组作为默认值，防止调用方收到 undefined
});
}

// 添加创建 ServiceAccount 的方法
export function createServiceAccount(data) {
  return request({
    url: '/create-sa',
    method: 'post',
    data:data
  })
}

export function getResources() {
  return request({
    url: '/resources',
    method: 'get'
  }).then(response => {
    console.log(response)
  // 校验后端返回的数据结构是否符合预期
  return response
})
.catch(error => {
  // 捕获并处理异常，避免未捕获的错误
  console.error('Error fetching sa:', error);
  return []; // 返回空数组作为默认值，防止调用方收到 undefined
});
}

export function updateServiceAccount(data) {
  return request({
    url: '/update-sa',
    method: 'post',
    data:data
  })
}

export function deleteSa(namespace, name) {
  return request({
    url: '/delete-sa',
    method: 'post',
    data: {
      namespace,
      name
    }
  });
}


export function listNodePool(data) {
  return request({
    url: '/nodepool/list',
    method: 'post',
    data:data
  })
}