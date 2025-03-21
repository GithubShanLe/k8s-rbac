<template>

  <el-card class="box-card">
        <!-- 节点列表 -->
    <el-card shadow="hover" class="mt-20 info-card">
      <div slot="header" class="sub-header">
        <i class="el-icon-cpu"></i> daemonset列表
        <span class="sub-title-count">({{ daemonsets.length || 0 }})</span>
      </div>
      <el-table 
        :data="daemonsets" 
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
          <el-table-column prop="pods" label="Pods" width="180" /> 
        <el-table-column prop="createTime" label="创建时间" width="180" />
      </el-table>
    </el-card>
  </el-card>
</template>

<script>

import { listDaemonset } from '@/api/k8s'
export default {
data() {
return {
  listLoading: false,
  daemonsets: [],
  labels: {},  // 初始化 labels 为空对象
  labelsFolded: true  // 添加折叠状态控制
}
},
mounted() {
this.fetchData()
},
methods: {
toggleLabels(index) {
  // 使用 Vue.set 确保响应式更新
  if (!this.daemonset[index].labelsExpanded) {
    this.$set(this.daemonset[index], 'labelsExpanded', true);
  } else {
    this.$set(this.daemonset[index], 'labelsExpanded', false);
  }
},

async fetchData() {
  this.listLoading = true
  try {
    const { data } = await listDaemonset({})
    if (data.errorCode) {
      this.$message.error(data.errorMessage || '获取数据失败')
      return
    }
    // 初始化每个服务的标签展开状态
    this.daemonsets = (data.daemonsets || []).map(daemonset => ({
      ...daemonset,
      labelsExpanded: false
    }))
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


