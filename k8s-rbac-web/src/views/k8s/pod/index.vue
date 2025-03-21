<template>

  <el-card class="box-card">
    <div class="filter-container">
      <el-select v-model="queryParams.nameSpace" placeholder="命名空间" class="filter-item" filterable  @change="handleNamespaceChange">
        <el-option
          v-for="ns in namespaces"
          :key="ns"
          :label="ns"
          :value="ns"
        />
      </el-select>
      <el-input
        v-model="queryParams.podName"
        placeholder="Pod名称"
        class="filter-item"
        style="width: 200px;"
        clearable
        @keyup.enter.native="fetchData"
      />
      <el-button class="filter-item" type="primary" @click="fetchData">
        查询
      </el-button>
    </div>
        <!-- 节点列表 -->
    <el-card shadow="hover" class="mt-20 info-card">
      <div slot="header" class="sub-header">
        <i class="el-icon-cpu"></i> pod列表
        <span class="sub-title-count">({{ pods.length || 0 }})</span>
      </div>
      
      <el-table 
        :data="pods" 
        border 
        stripe
        style="width: 100%">
        <el-table-column prop="name" label="名称"  min-width="30" show-overflow-tooltip/>
        <el-table-column prop="namespace" label="命名空间" width="180"  />  
        <el-table-column prop="images" label="镜像" min-width="30" show-overflow-tooltip />   
          <el-table-column label="标签" min-width="50" show-overflow-tooltip>
            <template slot-scope="{ row, $index }">
              <div class="tag-container">
                <template v-if="Object.keys(row.labels || {}).length > 0">
                  <!-- 始终显示前三个标签 -->
                  <div 
                    v-for="(value, key, index) in row.labels"
                    :key="key"
                    v-if="index < 3 || row.labelsExpanded"
                    class="tag-item-wrapper"
                  >
                    <el-tag
                      type="primary"
                      size="mini"
                      class="tag-item"
                    >
                      {{ key }}: {{ value }}
                    </el-tag>
                  </div>
                </template>
                <span v-else class="muted-text">无标签</span>
              </div>
              <div v-if="Object.keys(row.labels || {}).length > 3" class="fold-hint">
                <el-button type="text" size="mini" @click="toggleLabels($index)">
                  {{ row.labelsExpanded ? '收起' : `展开(${Object.keys(row.labels).length - 3}个)` }}
                  <i :class="row.labelsExpanded ? 'el-icon-arrow-up' : 'el-icon-arrow-down'"></i>
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="restart" label="Restart" width="180" /> 
          <el-table-column prop="cpu" label="cpu" width="180" /> 
          <el-table-column prop="mem" label="mem" width="180" />
          <el-table-column prop="cpu" label="Cpu使用率" width="180" />
          <el-table-column label="CPU限制率" width="120">
            <template slot-scope="{ row }">
              {{ row.metric && row.metric.cpuLimitRate !== undefined ? `${row.metric.cpuLimitRate}%` : 'N/A' }}
            </template>
          </el-table-column>
          <el-table-column label="内存限制率" width="120">
            <template slot-scope="{ row }">
              {{ row.metric && row.metric.memoryLimitRate !== undefined ? `${row.metric.memoryLimitRate}%` : 'N/A' }}
            </template>
          </el-table-column>
          
        <el-table-column prop="createTime" label="创建时间" width="180" />
      </el-table>
    </el-card>
  </el-card>
</template>

<script>

import { listPod, getPodMetrics, getNamespaces } from '@/api/k8s'  // 添加 listNamespace 导入
export default {
data() {
return {
  listLoading: false,
  pods: [],
  labels: {},  // 初始化 labels 为空对象
  labelsFolded: true,  // 添加折叠状态控制
  namespaces: [], // 添加命名空间列表
  queryParams: {
    nameSpace: localStorage.getItem('lastNamespace') || 'default',
    podName: ''
  }
}
},
async mounted() {
this.getNamespaces()
this.fetchData()
},
methods: {
toggleLabels(index) {
  // 使用 Vue.set 确保响应式更新
  if (!this.pods[index].labelsExpanded) {
    this.$set(this.pods[index], 'labelsExpanded', true);
  } else {
    this.$set(this.pods[index], 'labelsExpanded', false);
  }
},
handleNamespaceChange() {
      this.fetchData()
    },
async getNamespaces() {
  try {
    const res = await getNamespaces()
    console.log('原始返回数据:', res) // 查看完整返回数据
    
    if (!res) {
      console.warn('接口返回为空')
      this.namespaces = ['kube-system']
      return
    }

    if (res.items && Array.isArray(res.items)) {
      this.namespaces = res.items.map(item => item.metadata.name)
    } else if (Array.isArray(res)) {
      this.namespaces = res
    } else if (res.namespaces && Array.isArray(res.namespaces)) {
      this.namespaces = res.namespaces
    } else {
      console.warn('无法解析的数据格式:', res)
      this.namespaces = ['default']
    }
    
    console.log('处理后的命名空间列表:', this.namespaces)
  } catch (error) {
    console.error('获取命名空间列表失败:', error)
    this.namespaces = ['default']
  }
},

async fetchData() {
  this.listLoading = true
  try {
    localStorage.setItem('lastNamespace', this.queryParams.nameSpace)
    const { data } = await listPod({
      nameSpace: this.queryParams.nameSpace,
      podName: this.queryParams.podName
    })
    if (data.errorCode) {
      this.$message.error(data.errorMessage || '获取数据失败')
      return
    }
    
    // 初始化基础数据
    this.pods = (data.pods || []).map(pod => ({
      ...pod,
      labelsExpanded: false,
      metric: undefined  // 初始化为 undefined，表示数据未加载
    }))
    
    // 逐个获取指标数据
    this.pods.forEach(async (pod, index) => {
      try {
        const metricRes = await getPodMetrics({
          nameSpace: pod.namespace,
          podName: pod.name
        })
        if (metricRes.data && !metricRes.data.errorCode) {
          this.$set(this.pods[index], 'metric', metricRes.data.metric || {})
        }
      } catch (error) {
        console.error(`获取 ${pod.name} 指标失败:`, error)
        this.$set(this.pods[index], 'metric', {})  // 失败时设置为空对象
      }
    })
    
  } catch (error) {
    this.$message.error('请求异常：' + error.message)
  } finally {
    this.listLoading = false
  }
}
}
}
</script>

<style scoped>
.box-card {
margin: 20px;
}
.fr {
float: right;
}
.mt-20 {
margin-top: 20px;
}
.info-card {
margin-bottom: 0;
}
.sub-header {
font-size: 16px;
font-weight: bold;
color: #303133;
display: flex;
align-items: center;
}
.sub-header i {
margin-right: 8px;
font-size: 18px;
}
.sub-title-count {
margin-left: 8px;
font-size: 14px;
color: #909399;
font-weight: normal;
}
.info-item {
display: flex;
margin-bottom: 15px;
align-items: center;
}
.info-item .label {
font-weight: bold;
width: 100px;
color: #606266;
}
.info-item .value {
flex: 1;
color: #303133;
}
.label-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.filter-container {
  padding-bottom: 10px;
  display: flex;
  align-items: center;
}

.filter-item {
  margin-right: 10px;
}
</style>
.tag-container {
  display: flex;
  flex-direction: column; /* 改为垂直布局 */
  gap: 5px;
}

.tag-item-wrapper {
  width: 100%;
}

.tag-item {
  width: 100%; /* 标签占满容器宽度 */
  margin: 10; /* 移除默认边距 */
  justify-content: flex-start; /* 左对齐文本 */
}
</style>


