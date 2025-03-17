<template>
  <div style="padding:30px;">
    <el-alert :closable="false" title="RBAC Management" />
    <!-- Namespace 选择器 -->
    <el-select
      v-model="selectedNamespace"
      style="margin-top: 20px; width: 300px"
      placeholder="请选择Namespace"
      filterable
      default-first-option
      :filter-method="filterNamespaces"
      @change="fetchServiceAccounts"
    >
      <el-option
        v-for="ns in filteredNamespaces"
        :key="ns"
        :label="ns"
        :value="ns"
      />
    </el-select>
    <CreateServiceAccount 
    ref="createServiceAccountRef"
    :namespaces="namespaces"
    :edit-mode="editMode"
    :edit-data="editData"
    @created="handleSaUpdated" 
    @updated="handleSaUpdated"
  />
    <!-- Service Account 列表 -->
    <div class="sa-list" v-loading="loading">
      <div 
        v-for="sa in serviceAccounts" 
        :key="sa.name" 
        class="sa-card"
      >
        <div class="sa-info" @click="handleSaClick(sa)">
          <div class="sa-name">{{ sa.name }}</div>
          <div class="sa-namespace">{{ sa.namespace }}</div>
        </div>
        <div class="sa-actions">
          <el-button
            type="text"
            class="delete-btn"
            @click.stop="handleDelete(sa)"
          >
            <i class="el-icon-delete"></i>
          </el-button>
          <i class="el-icon-arrow-right"></i>
        </div>
      </div>
    </div>

    <!-- SA 详情对话框 -->
    <el-dialog
      :title="'Service Account: ' + (currentSa ? currentSa.name : '')"
      :visible.sync="dialogVisible"
      width="90%"
      append-to-body
    >
      <div v-if="currentSa" class="sa-detail">
        <div class="action-buttons" style="margin-bottom: 20px;">
          <el-button 
            type="warning" 
            @click="handleEdit"
          >
            编辑权限
          </el-button>
        </div>

        <h4>基本信息</h4>
        <div class="basic-info">
          <div class="info-item">
            <div class="info-label">名称:</div>
            <div class="info-value">{{ currentSa.name }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">命名空间:</div>
            <div class="info-value">{{ currentSa.namespace }}</div>
          </div>
        </div>

        <h4>角色权限</h4>
        <div class="permission-summary">
          <p>该 ServiceAccount 拥有以下角色和权限，点击"编辑权限"按钮可以修改权限。</p>
          <ul>
            <li v-for="role in saRoles" :key="role.roleName">
              角色 <strong>{{ role.roleName }}</strong> ({{ role.rules.length }} 条规则)
            </li>
            <li v-for="role in saClusterRoles" :key="role.clusterRoleName">
              集群角色 <strong>{{ role.clusterRoleName }}</strong> ({{ role.rules.length }} 条规则)
            </li>
          </ul>
        </div>
      </div>
    </el-dialog>

    <!-- YAML 预览对话框 -->
    <el-dialog
      title="YAML 预览"
      :visible.sync="yamlDialogVisible"
      append-to-body
      width="80%"
    >
      <pre style="background-color: #f5f7fa; padding: 15px; border-radius: 4px; max-height: 600px; overflow: auto;">{{ generatedYAML }}</pre>
      <span slot="footer" class="dialog-footer">
        <el-button @click="yamlDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleCopyYAML">复制</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getNamespaces, getServiceAccounts, getSaRoles,deleteSa } from '@/api/k8s'
import CreateServiceAccount from './components/CreateServiceAccount.vue'  // 添加这行导入

export default {
  components: {
    CreateServiceAccount  // 注册组件
  },
  data() {
    return {
      selectedNamespace: '',
      namespaces: [],
      filteredNamespaces: [],
      serviceAccounts: [],
      loading: false,
      dialogVisible: false,
      saRoles: [],
      saClusterRoles: [],
      currentSa: null,
      allVerbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],
      yamlDialogVisible: false,
      generatedYAML: '',
      editMode: false,
      editData: null
    }
  },

  created() {
    this.fetchNamespaces()
  },

  methods: {
      // 处理编辑按钮点击
    handleEdit() {
      // 准备编辑数据
      this.prepareEditData();
      // 关闭详情对话框
      this.dialogVisible = false;
      // 打开编辑对话框
      this.$nextTick(() => {
        this.editMode = true;
        this.$refs.createServiceAccountRef.showDialog = true;
      });
    },
    // 准备编辑数据
    prepareEditData() {
      if (!this.currentSa || !this.saRoles || !this.saClusterRoles) return;

      // 将当前 SA 详情转换为 CreateServiceAccount 组件需要的格式
      this.editData = {
        name: this.currentSa.name,
        namespace: this.currentSa.namespace,
        roles: this.saRoles.map(role => ({
          roleName: role.roleName,
          roleBindingName: role.roleBindingName,
          nameSpace: role.nameSpace,
          rules: role.rules.map(rule => ({
            resource: rule.resource,
            apiGroup: Array.isArray(rule.apiGroup) ? rule.apiGroup : [rule.apiGroup],
            verbs: rule.verbs.includes('*') ? this.allVerbs : rule.verbs
          }))
        })),
        clusterRoles: this.saClusterRoles.map(role => ({
          clusterRoleName: role.clusterRoleName,
          clusterRoleBindingName: role.clusterRoleBindingName,
          nameSpace: role.nameSpace,
          rules: role.rules.map(rule => ({
            resource: rule.resource,
            apiGroup: Array.isArray(rule.apiGroup) ? rule.apiGroup : [rule.apiGroup],
            verbs:    rule.verbs.includes('*') ? this.allVerbs : rule.verbs

          }))
        }))
      };
    },

    // 处理 SA 更新后的回调
    async handleSaUpdated() {
      // 重置编辑模式
      this.editMode = false;
      this.editData = null;
      
      // 刷新 ServiceAccount 列表
      await this.fetchServiceAccounts();
      
      // 如果当前有选中的 SA，刷新其详情
      if (this.currentSa) {
        await this.fetchSaRoles(this.currentSa);
      }
      
      this.$message.success('ServiceAccount 更新成功');
    },
    async fetchNamespaces() {
      try {
        this.loading = true
        const data = await getNamespaces()
        this.namespaces = data
        this.filteredNamespaces = this.namespaces
      } catch (error) {
        console.error('Failed to fetch namespaces:', error)
        this.namespaces = []
        this.filteredNamespaces = []
      } finally {
        this.loading = false
      }
    },

    async fetchServiceAccounts() {
      if (!this.selectedNamespace) return
      try {
        this.loading = true
        const response = await getServiceAccounts(this.selectedNamespace)
        this.serviceAccounts = response.data.serviceAccounts || []
      } catch (error) {
        console.error('Failed to fetch service accounts:', error)
        this.serviceAccounts = []
      } finally {
        this.loading = false
      }
    },

    filterNamespaces(query) {
      if (query) {
        this.filteredNamespaces = this.namespaces.filter(ns => 
          ns.toLowerCase().includes(query.toLowerCase())
        )
      } else {
        this.filteredNamespaces = this.namespaces
      }
    },

    async handleSaClick(row) {
      this.currentSa = row
      this.dialogVisible = true
      await this.fetchSaRoles(row)
    },

    async fetchSaRoles(sa) {
      try {
        this.loading = true
        const response = await getSaRoles(sa.namespace, sa.name)
        const info = response.data.serviceAccountInfo
        this.saRoles = info.roles || []
        this.saClusterRoles = info.clusterRoles || []
      } catch (error) {
        console.error('Failed to fetch SA roles:', error)
        this.saRoles = []
        this.saClusterRoles = []
      } finally {
        this.loading = false
      }
    },

    handleVerbChange(checked, verb, rule) {
      if (!rule.verbs) {
        rule.verbs = []
      }

      // 如果原来是 '*'，需要展开成所有权限
      if (rule.verbs.includes('*')) {
        rule.verbs = [...this.allVerbs]
      }
      
      if (checked && !rule.verbs.includes(verb)) {
        rule.verbs.push(verb)
        // 如果选中后包含所有权限，转换为 '*'
        if (this.allVerbs.every(v => rule.verbs.includes(v))) {
          rule.verbs = ['*']
        }
      } else if (!checked) {
        // 取消选中时，移除该动词
        rule.verbs = rule.verbs.filter(v => v !== verb)
      }
    },

    handleGenerateYAML() {
      const roleRules = this.saRoles.map(role => ({
        name: role.roleName,
        namespace: role.nameSpace,
        bindingName: role.roleBindingName,
        rules: role.rules.map(rule => {
          // 如果是 '*'，需要展开所有权限用于检查
          const currentVerbs = rule.verbs.includes('*') ? [...this.allVerbs] : rule.verbs

          return {
            apiGroups: rule.apiGroup,
            resources: [rule.resource],
            verbs: this.allVerbs.every(verb => currentVerbs.includes(verb)) ? 
              ['*'] : 
              [...currentVerbs]
          }
        })
      }))

      // 按照界面上显示的 clusterRole 状态组织规则
      const clusterRoleRules = this.saClusterRoles.map(role => ({
        name: role.clusterRoleName,
        bindingName: role.clusterRoleBindingName,
        rules: role.rules.map(rule => {
          // 如果是 '*'，需要展开所有权限用于检查
          const currentVerbs = rule.verbs.includes('*') ? [...this.allVerbs] : rule.verbs

          return {
            apiGroups: rule.apiGroup,
            resources: [rule.resource],
            verbs: this.allVerbs.every(verb => currentVerbs.includes(verb)) ? 
              ['*'] : 
              [...currentVerbs]
          }
        })
      }))

      this.generatedYAML = this.generateYAML(roleRules, clusterRoleRules)
      this.yamlDialogVisible = true
    },

    generateYAML(roleRules, clusterRoleRules) {
      if (!this.currentSa) return ''

      return `apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${this.currentSa.name}
  namespace: ${this.currentSa.namespace}
---
${roleRules.map(role => `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ${role.name}
  namespace: ${role.namespace}
rules:
${role.rules.map(rule => `  - apiGroups: ${JSON.stringify(rule.apiGroups)}
    resources: ${JSON.stringify(rule.resources)}
    verbs: ${JSON.stringify(rule.verbs)}`).join('\n')}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${role.bindingName}
  namespace: ${role.namespace}
subjects:
- kind: ServiceAccount
  name: ${this.currentSa.name}
  namespace: ${this.currentSa.namespace}
roleRef:
  kind: Role
  name: ${role.name}
  apiGroup: rbac.authorization.k8s.io`).join('\n---\n')}${clusterRoleRules.length > 0 ? '\n---\n' : ''}${clusterRoleRules.map(role => `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ${role.name}
rules:
${role.rules.map(rule => `  - apiGroups: ${JSON.stringify(rule.apiGroups)}
    resources: ${JSON.stringify(rule.resources)}
    verbs: ${JSON.stringify(rule.verbs)}`).join('\n')}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${role.bindingName}
subjects:
- kind: ServiceAccount
  name: ${this.currentSa.name}
  namespace: ${this.currentSa.namespace}
roleRef:
  kind: ClusterRole
  name: ${role.name}
  apiGroup: rbac.authorization.k8s.io`).join('\n---\n')}`
    },

    handleCopyYAML() {
      navigator.clipboard.writeText(this.generatedYAML)
        .then(() => {
          this.$message.success('YAML 已复制到剪贴板')
        })
        .catch(() => {
          this.$message.error('复制失败')
        })
    },
    async handleDelete(sa) {
      try {
        await this.$confirm('确认删除该 ServiceAccount 吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        });
        
        this.loading = true;
        await deleteSa(sa.namespace, sa.name);
        this.$message.success('删除成功');
        await this.fetchServiceAccounts();
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除 ServiceAccount 失败:', error);
          this.$message.error('删除失败');
        }
      } finally {
        this.loading = false;
      }
    },
  }
}
</script>

<style scoped>
.rule-item {
  margin-bottom: 10px;
  padding: 5px;
  border-bottom: 1px solid #eee;
}
.resource-group {
  margin-bottom: 5px;
}
.resource-name {
  font-weight: bold;
  margin-right: 5px;
}
.api-group {
  color: #909399;
  font-size: 12px;
}
.verbs-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 5px;
}
.el-checkbox {
  margin-right: 15px;
  margin-left: 0;
}
.el-table__expanded-cell {
  padding: 20px 50px !important;
}
.el-table__expanded-cell[class*=cell] {
  padding: 20px 50px !important;
}

.role-section {
  margin-bottom: 20px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.1);
}

.role-info {
  margin-bottom: 10px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 4px 4px 0 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-label {
  color: #606266;
  font-weight: 600;
  min-width: 100px;
}

.info-value {
  color: #303133;
  font-family: monospace;
  background-color: #fff;
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
  flex: 1;
}
</style>

<style scoped>
.rules-container {
  border: 1px solid #EBEEF5;
  border-radius: 4px;
  margin: 15px 0;
  background: #fff;
}

.rule-header {
  display: flex;
  background-color: #f5f7fa;
  padding: 12px 20px;
  border-bottom: 2px solid #DCDFE6;
}

.rule-row {
  display: flex;
  align-items: flex-start;
  padding: 16px 20px;
  border-bottom: 1px solid #EBEEF5;
  transition: background-color 0.3s;
}

.rule-row:hover {
  background-color: #f5f7fa;
}

.resource-info {
  width: 250px;
  flex-shrink: 0;
  padding-right: 20px;
}

.api-info {
  width: 400px;
  flex-shrink: 0;
  padding: 0 20px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.verbs-info {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding-left: 20px;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.api-tag {
  font-family: monospace;
}

.el-checkbox {
  margin: 0;
  min-width: 80px;
}

.el-tag + .el-tag {
  margin-left: 4px;
}
</style>

<style scoped>
.sa-list {
  margin-top: 20px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.sa-card {
  background: #fff;
  border-radius: 4px;
  padding: 16px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid #EBEEF5;
}

.sa-info {
  flex: 1;
  cursor: pointer;
}

.sa-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.delete-btn {
  padding: 0;
  color: #F56C6C;
}

.delete-btn:hover {
  color: #f78989;
}

.el-icon-arrow-right {
  color: #909399;
  font-size: 18px;
}
</style>

<style scoped>
.basic-info {
  background-color: #f5f7fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.permission-summary {
  background-color: #f5f7fa;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
}

.permission-summary ul {
  margin-top: 10px;
  padding-left: 20px;
}

.permission-summary li {
  margin-bottom: 8px;
}
</style>