<template>
  <div class="app-container">
    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>节点池详情 - {{ poolInfo.name }}</span>
        <el-button type="primary" icon="el-icon-back" class="fr" @click="goBack">返回</el-button>
      </div>
      
      <!-- 基本信息卡片 -->
      <el-row :gutter="20">
        <el-col :span="24">
          <el-card shadow="hover" class="info-card">
            <div slot="header" class="sub-header">
              <i class="el-icon-info"></i> 基本信息
            </div>
            <div class="basic-info-container">
                <div class="info-block">
                  <div class="info-icon">
                    <i class="el-icon-s-flag"></i>
                  </div>
                  <div class="info-content">
                    <div class="info-label">节点池名称</div>
                    <div class="info-value">{{ poolInfo.name }}</div>
                  </div>
                </div>
                
                <div class="info-block">
                  <div class="info-icon">
                    <i class="el-icon-s-operation"></i>
                  </div>
                  <div class="info-content">
                    <div class="info-label">状态</div>
                    <div class="info-value">
                      <el-tag :type="statusType(poolInfo.status)">{{ poolInfo.status }}</el-tag>
                    </div>
                  </div>
                </div>
                
                <div class="info-block">
                  <div class="info-icon">
                    <i class="el-icon-s-grid"></i>
                  </div>
                  <div class="info-content">
                    <div class="info-label">节点数量</div>
                    <div class="info-value">{{ poolInfo.nodeCounts }}</div>
                  </div>
                </div>
              </div>
          </el-card>
        </el-col>
      </el-row>
      
      <!-- 标签和污点信息 -->
      <el-row :gutter="20" class="mt-20">
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="info-card">
            <div slot="header" class="sub-header">
              <i class="el-icon-collection-tag"></i> 标签
              <el-button 
                type="text" 
                class="fold-button" 
                @click="labelsFolded = !labelsFolded">
                {{ labelsFolded ? '展开' : '收起' }}
                <i :class="labelsFolded ? 'el-icon-arrow-down' : 'el-icon-arrow-up'"></i>
              </el-button>
            </div>
            <div class="tag-container" :class="{'is-folded': labelsFolded}">
              <template v-if="Object.keys(poolInfo.labels || {}).length > 0">
                <el-tag
                  v-for="(value, key) in poolInfo.labels"
                  :key="key"
                  type="info"
                  size="medium"
                  class="tag-item"
                >
                  {{ key }}: {{ value }}
                </el-tag>
              </template>
              <span v-else class="muted-text">无标签</span>
            </div>
            <div v-if="labelsFolded && Object.keys(poolInfo.labels || {}).length > 3" class="fold-hint">
              还有 {{ Object.keys(poolInfo.labels || {}).length - 3 }} 个标签...
            </div>
          </el-card>
        </el-col>
        
        <el-col :xs="24" :sm="12">
          <el-card shadow="hover" class="info-card">
            <div slot="header" class="sub-header">
              <i class="el-icon-warning"></i> 污点
              <el-button 
                type="text" 
                class="fold-button" 
                @click="taintsFolded = !taintsFolded">
                {{ taintsFolded ? '展开' : '收起' }}
                <i :class="taintsFolded ? 'el-icon-arrow-down' : 'el-icon-arrow-up'"></i>
              </el-button>
            </div>
            <div class="tag-container" :class="{'is-folded': taintsFolded}">
              <template v-if="Object.keys(poolInfo.taints || {}).length > 0">
                <el-tag
                  v-for="(value, key) in poolInfo.taints"
                  :key="key"
                  type="danger"
                  size="medium"
                  class="tag-item"
                >
                  {{ key }}: {{ value }}
                </el-tag>
              </template>
              <span v-else class="muted-text">无污点</span>
            </div>
            <div v-if="taintsFolded && Object.keys(poolInfo.taints || {}).length > 3" class="fold-hint">
              还有 {{ Object.keys(poolInfo.taints || {}).length - 3 }} 个污点...
            </div>
          </el-card>
        </el-col>
      </el-row>
      
      <!-- 节点列表 -->
      <el-card shadow="hover" class="mt-20 info-card">
        <div slot="header" class="sub-header">
          <i class="el-icon-cpu"></i> 节点列表
          <span class="sub-title-count">({{ poolInfo.nodeCounts || 0 }})</span>
        </div>
        <el-table 
          :data="poolInfo.nodeList" 
          border 
          stripe
          style="width: 100%">
          <el-table-column prop="name" label="节点名称" min-width="50" show-overflow-tooltip />
          <el-table-column prop="nodeIp" label="IP" width="140" />
          <el-table-column label="CPU(R/L)" width="200">
            <template slot-scope="scope">
              <el-progress 
                :percentage="calculateCpuUsage(scope.row)" 
                :color="getResourceColor(calculateCpuUsage(scope.row))"
                :format="format => `${(scope.row.requestCpu/1000).toFixed(1)}/${(scope.row.limitCpu/1000).toFixed(1)} C`"
              ></el-progress>
            </template>
          </el-table-column>
          <el-table-column label="内存(R/L)" width="200">
            <template slot-scope="scope">
              <el-progress 
                :percentage="calculateMemUsage(scope.row)" 
                :color="getResourceColor(calculateMemUsage(scope.row))"
                :format="format => `${(scope.row.requestMem/1024/1024).toFixed(1)}/${(scope.row.limitMem/1024/1024).toFixed(1)} GiB`"
              ></el-progress>
            </template>
          </el-table-column>
          <el-table-column label="Pod(R/L)" width="140">
            <template slot-scope="scope">
              {{ scope.row.requestPod || 0 }}/{{ scope.row.limitPod || 0 }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="120">
            <template slot-scope="scope">
              <el-tag :type="statusType(scope.row.status)" size="mini">{{ scope.row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180" />
        </el-table>
      </el-card>
    </el-card>
  </div>
</template>

<script>
export default {
  data() {
    return {
      poolInfo: {},
      labelsFolded: true,
      taintsFolded: true
    }
  },
  created() {
    const poolName = this.$route.query.name
    if (!poolName) {
      this.$message.error('节点池名称不存在')
      this.goBack()
      return
    }

    try {
      const poolInfoStr = sessionStorage.getItem('nodepool_' + poolName)
      if (poolInfoStr) {
        this.poolInfo = JSON.parse(poolInfoStr)
      } else {
        this.$message.error('节点池信息不存在')
        this.goBack()
      }
    } catch (error) {
      console.error('解析节点池信息失败:', error)
      this.$message.error('解析节点池信息失败')
      this.goBack()
    }
  },
  methods: {
    goBack() {
       // 修改为 replace 模式导航
       if (window.history.length > 1) {
            this.$router.replace('/kubernetes/nodepool') // 替换当前历史记录
        } else {
            this.$router.push('/kubernetes/nodepool')
        }
    },
    statusType(status) {
      const typeMap = {
        'Active': 'success',
        'Error': 'danger',
        'Ready': 'primary'
      }
      return typeMap[status] || 'warning'
    },
    // 计算CPU使用率
    calculateCpuUsage(node) {
      if (!node.limitCpu || node.limitCpu === 0) return 0
      return Math.min(100, Math.round((node.requestCpu / node.limitCpu) * 100))
    },
    // 计算内存使用率
    calculateMemUsage(node) {
      if (!node.limitMem || node.limitMem === 0) return 0
      return Math.min(100, Math.round((node.requestMem / node.limitMem) * 100))
    },
    // 根据使用率返回颜色
    getResourceColor(percentage) {
      if (percentage < 70) return '#67C23A'
      if (percentage < 90) return '#E6A23C'
      return '#F56C6C'
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
.tag-container {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.tag-item {
  margin: 2px;
}
.muted-text {
  color: #909399;
  font-style: italic;
}
.fold-button {
  margin-left: auto;
  font-size: 13px;
}

.tag-container.is-folded {
  max-height: 40px;
  overflow: hidden;
}

.fold-hint {
  color: #909399;
  font-size: 12px;
  margin-top: 8px;
  text-align: right;
}

.info-icon {
    font-size: 24px;
    color: #409EFF;
    margin-right: 15px;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: rgba(64, 158, 255, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .basic-info-container {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-around;
    padding: 10px 0;
  }

  .info-block {
    display: flex;
    align-items: center;
    padding: 15px;
    min-width: 200px;
    flex: 1;
    border-radius: 8px;
    background-color: #f8f9fa;
    margin: 0 10px;
    transition: all 0.3s;
  }

  .info-block:hover {
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .info-content {
    flex: 1;
  }
  
  .info-label {
    font-size: 13px;
    color: #909399;
    margin-bottom: 5px;
  }
  
  .info-value {
    font-size: 16px;
    color: #303133;
    font-weight: bold;
  }
  
</style>


