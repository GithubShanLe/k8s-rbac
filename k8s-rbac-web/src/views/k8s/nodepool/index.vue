<template>
  <div class="app-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>节点池管理</span>
        <el-button type="primary" icon="el-icon-search" class="fr">筛选</el-button>
      </div>
      
      <el-row :gutter="20" v-loading="listLoading">
        <el-col 
          v-for="pool in tableData" 
          :key="pool.name"
          :xs="24" :sm="12" :md="8" :lg="6"
          class="pool-col">
          <el-card 
            class="pool-card" 
            @click.native="goToNodeDetail(pool)"
            :body-style="{ padding: '15px', cursor: 'pointer' }">
            <div class="pool-header">
              <h3 class="pool-name">{{ pool.name }}</h3>
              <el-tag :type="statusType(pool.status)" size="mini">
                {{ pool.status }}
              </el-tag>
            </div>
            
            <div class="pool-meta">
              <span class="meta-item">节点数：{{ pool.nodeCounts }}</span>
            </div>
            
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script>
import { listNodePool } from '@/api/k8s'

export default {
  data() {
    return {
      listLoading: false,
      tableData: []
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    goToNodeDetail(pool) {
      try {
        // 先存储数据，避免路由跳转后数据丢失
        const poolData = JSON.stringify(pool)
        sessionStorage.setItem('nodepool_' + pool.name, poolData)
        
        // 使用 query 参数传递简单数据
        this.$router.push({
          path: '/k8s/nodepool/detail',
          query: { name: pool.name }
        })
      } catch (error) {
        console.error('跳转错误:', error)
        this.$message.error('页面跳转失败: ' + error.message)
      }
    },
    async fetchData() {
      this.listLoading = true
      try {
        const { data } = await listNodePool({})
        if (data.errorCode) {
          this.$message.error(data.errorMessage || '获取数据失败')
          return
        }
        this.tableData = data.nodePools.map(pool => ({
          name: pool.name,
          status: pool.status,
          labels: pool.lables,
          taints: pool.taints,
          nodeCounts: pool.nodeList.length,
          nodeList: pool.nodeList,
          showNodes: false // 添加展开状态控制
        }))
      } catch (error) {
        this.$message.error('请求异常：' + error.message)
      } finally {
        this.listLoading = false
      }
    },
    statusType(status) {
      const typeMap = {
        'Active': 'success',
        'Error': 'danger',
        'Ready': 'primary'
      }
      return typeMap[status] || 'warning'
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
.tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  position: relative;
  max-height: 32px;
  overflow: hidden;
  line-height: 1.5;
}
.tag-group::after {
  content: "";
  position: absolute;
  bottom: 0;
  right: 0;
  left: 0;
  height: 10px;
  background: linear-gradient(to bottom, rgba(255,255,255,0), rgba(255,255,255,1) 80%);
}
.tag-item {
  font-family: Monaco, Consolas, monospace;
  font-size: 12px;
  background-color: #ecf5ff;
  border-color: #d9ecff;
  color: #409eff;
  margin: 2px 0;
}
.muted-text {
  color: #909399;
  font-style: italic;
}
.app-container {
  height: calc(100vh - 84px);
}
.box-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.pool-col {
  margin-bottom: 20px;
}
.pool-card {
  height: 100%;
  transition: transform 0.2s;
}
.pool-card:hover {
  transform: translateY(-3px);
}
.pool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.pool-name {
  margin: 0;
  font-size: 16px;
  color: #303133;
}
.pool-meta {
  margin: 10px 0;
  font-size: 13px;
  color: #606266;
}

/* 新增展开指示器样式 */
.expand-indicator {
  margin-top: 12px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.expand-indicator:hover {
  background: #f0f2f5;
}
.is-expanded {
  transform: rotate(180deg);
  transition: transform 0.3s;
}
.node-list {
  margin-top: 12px;
  border-top: 1px solid #ebeef5;
  padding-top: 12px;
}
</style>

<style>
.el-table__row:hover {
  cursor: pointer;
  background-color: #f5f7fa;
}
</style>