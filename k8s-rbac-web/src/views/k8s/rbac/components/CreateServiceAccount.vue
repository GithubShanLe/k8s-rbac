<template>
  <div>
    <el-button
      v-if="!editMode"
      type="primary"
      @click="showDialog = true"
    >
      创建 ServiceAccount
    </el-button>
    <el-dialog
      :title="editMode ? '编辑 ServiceAccount' : '创建 ServiceAccount'"
      :visible.sync="showDialog"
      width="80%"
      @open="handleDialogOpen"
    >

      <el-form :model="form" label-width="120px">
        <!-- 添加 YAML 导入功能 -->
        <el-form-item label="导入 YAML">
          <el-input
            type="textarea"
            :rows="5"
            v-model="importYaml"
            placeholder="粘贴 YAML 配置..."
          ></el-input>
          <el-button type="primary" size="small" @click="parseYaml" style="margin-top: 10px;">解析 YAML</el-button>
        </el-form-item>      
      </el-form>
      
      <el-form :model="form" label-width="120px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item label="命名空间" required>
          <el-select
            v-model="form.namespace"
            filterable
            allow-create
            remote
            :remote-method="filterNamespace"
            placeholder="请选择或输入命名空间"
          >
            <el-option
              v-for="ns in filteredNamespaces"
              :key="ns"
              :label="ns"
              :value="ns"
            />
          </el-select>
        </el-form-item>

        <!-- 角色配置 -->
        <el-form-item label="角色">
          <el-button type="text" @click="addRole">添加角色</el-button>
          <div v-for="(role, index) in form.roles" :key="index" class="role-section">
            <el-form-item label="角色名称">
              <el-input v-model="role.roleName" @input="updateRoleBindingName(role)"></el-input>
            </el-form-item>
            <el-form-item label="角色绑定名称">
              <el-input v-model="role.roleBindingName"></el-input>
            </el-form-item>
            <el-form-item label="规则">
              <el-button type="text" @click="addRule(role)">添加规则</el-button>
              <div class="rules-container">
                <div class="rule-header">
                  <div class="resource-info">
                    <span class="header-title">Resource</span>
                  </div>
                  <div class="api-info">
                    <span class="header-title">API Group</span>
                  </div>
                  <div class="verbs-info">
                    <span class="header-title">Permissions</span>
                  </div>
                </div>
                <div v-for="(rule, ruleIndex) in role.rules" :key="ruleIndex" class="rule-row">
                  <div class="resource-info">
                    <el-select
                      v-model="rule.resource"
                      filterable
                      placeholder="请选择资源"
                      @focus="loadResources"
                      @change="(val) => handleResourceChange(val, rule)"
                    >
                      <el-option
                        v-for="resource in filteredResources(true)"
                        :key="resource.Kind"
                        :label="resource.Kind"
                        :value="resource.Kind"
                      />
                    </el-select>
                  </div>
                  <div class="api-info">
                    <div v-if="rule.resource && (isWildcardResource(rule.resource) || hasMultipleApiGroups(rule.resource))">
                      <el-select
                        v-model="rule.apiGroup"
                        multiple
                        filterable
                        placeholder="请选择API Group"
                        class="api-group-select"
                      >
                        <el-option
                          v-for="group in getAvailableApiGroups(rule.resource)"
                          :key="group"
                          :label="group || ''"
                          :value="group"
                        />
                      </el-select>
                    </div>
                    <div v-else-if="rule.resource">
                      <el-tag
                        v-for="group in getApiGroupsForResource(rule.resource)"
                        :key="group"
                        type="success"
                        size="small"
                        class="api-tag"
                      >
                        {{ group || '' }}
                      </el-tag>
                    </div>
                    <span v-else class="placeholder">请先选择资源</span>
                  </div>
                  <div class="verbs-info">
                    <el-checkbox-group v-model="rule.verbs">
                      <el-checkbox
                        v-for="verb in allVerbs"
                        :key="verb"
                        :label="verb"
                      >
                        {{ verb }}
                      </el-checkbox>
                    </el-checkbox-group>
                  </div>
                  <el-button type="danger" size="mini" @click="removeRule(role, ruleIndex)">删除规则</el-button>
                </div>
              </div>
            </el-form-item>
            <el-button type="danger" size="mini" @click="removeRole(index)">删除角色</el-button>
          </div>
        </el-form-item>

        <!-- 集群角色配置 -->
        <el-form-item label="集群角色">
          <el-button type="text" @click="addClusterRole">添加集群角色</el-button>
          <div v-for="(clusterRole, index) in form.clusterRoles" :key="index" class="role-section">
            <el-form-item label="集群角色名称">
              <el-input v-model="clusterRole.clusterRoleName" @input="updateClusterRoleBindingName(clusterRole)"></el-input>
            </el-form-item>
            <el-form-item label="集群角色绑定名称">
              <el-input v-model="clusterRole.clusterRoleBindingName"></el-input>
            </el-form-item>
            <el-form-item label="规则">
              <el-button type="text" @click="addClusterRule(clusterRole)">添加规则</el-button>
              <div class="rules-container">
                <div class="rule-header">
                  <div class="resource-info">
                    <span class="header-title">Resource</span>
                  </div>
                  <div class="api-info">
                    <span class="header-title">API Group</span>
                  </div>
                  <div class="verbs-info">
                    <span class="header-title">Permissions</span>
                  </div>
                </div>
                <div v-for="(rule, ruleIndex) in clusterRole.rules" :key="ruleIndex" class="rule-row">
                  <div class="resource-info">
                    <el-select
                      v-model="rule.resource"
                      filterable
                      placeholder="请选择资源"
                      @focus="loadResources"
                      @change="(val) => handleResourceChange(val, rule)"
                    >
                      <el-option
                        v-for="resource in filteredResources(false)"
                        :key="resource.Kind"
                        :label="resource.Kind"
                        :value="resource.Kind"
                      />
                    </el-select>
                  </div>
                  <div class="api-info">
                    <div v-if="rule.resource && (isWildcardResource(rule.resource) || hasMultipleApiGroups(rule.resource))">
                      <el-select
                        v-model="rule.apiGroup"
                        multiple
                        filterable
                        placeholder="请选择API Group"
                        class="api-group-select"
                      >
                        <el-option
                          v-for="group in getAvailableApiGroups(rule.resource)"
                          :key="group"
                          :label="group || ''"
                          :value="group"
                        />
                      </el-select>
                    </div>
                    <div v-else-if="rule.resource">
                      <el-tag
                        v-for="group in getApiGroupsForResource(rule.resource)"
                        :key="group"
                        type="success"
                        size="small"
                        class="api-tag"
                      >
                        {{ group || '' }}
                      </el-tag>
                    </div>
                    <span v-else class="placeholder">请先选择资源</span>
                  </div>
                  <div class="verbs-info">
                    <el-checkbox-group v-model="rule.verbs">
                      <el-checkbox
                        v-for="verb in allVerbs"
                        :key="verb"
                        :label="verb"
                      >
                        {{ verb }}
                      </el-checkbox>
                    </el-checkbox-group>
                  </div>
                  <!-- 添加删除规则按钮 -->
                  <el-button type="danger" size="mini" @click="removeClusterRule(clusterRole, ruleIndex)">删除规则</el-button>
                </div>
              </div>
            </el-form-item>
            <!-- 将删除规则按钮移到外面 -->
            <el-button type="danger" size="mini" @click="removeClusterRole(index)">删除集群角色</el-button>
          </div>
        </el-form-item>
        <!-- YAML 预览区域 -->
        <el-form-item label="YAML 预览">
          <pre class="yaml-preview">{{ yamlConfig }}</pre>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="editMode ? handleUpdate() : handleCreate()">
          {{ editMode ? '更新' : '创建' }}
        </el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getResources, createServiceAccount,updateServiceAccount } from '@/api/k8s'
import yaml from 'js-yaml'

export default {
  props: {
    namespaces: {
      type: Array,
      required: true
    },
    editMode: {
      type: Boolean,
      default: false
    },
    editData: {
      type: Object,
      default: null
    }
  },
  data() {
    return {
      showDialog: false,
      importYaml: '', // 新增：用于存储导入的 YAML 文本
      form: {
        name: '',
        namespace: '',
        roles: [],
        clusterRoles: []
      },
      apiGroups: [''],
      allVerbs: ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete'],
      filteredNamespaces: [],
      allResources: [],
    }
  },
  computed: {
    yamlConfig() {
      const configs = [];
      
      if (!this.form.name || !this.form.namespace) {
        return '';
      }
  
      // ServiceAccount
      configs.push(yaml.dump({
        apiVersion: 'v1',
        kind: 'ServiceAccount',
        metadata: {
          name: this.form.name,
          namespace: this.form.namespace
        }
      }));
  
      // Roles
      this.form.roles.forEach(role => {
        if (role.roleName && role.rules.length > 0) {
          // Role
          const roleRules = role.rules
            .filter(rule => rule.resource && rule.verbs.length > 0)
            .map(rule => {
              const resource = rule.resource;
              const verbs = rule.verbs.length === this.allVerbs.length ? ['*'] : rule.verbs;
              // 修改这里：确保 apiGroups 是一维数组而不是嵌套数组
              let apiGroups;
              if (this.isWildcardResource(rule.resource) || this.hasMultipleApiGroups(rule.resource)) {
                // 将 rule.apiGroup 展平为一维数组
                apiGroups = Array.isArray(rule.apiGroup) ? 
                  rule.apiGroup.flat() : 
                  [rule.apiGroup];
              } else {
                // 获取资源的 API 组并确保是一维数组
                const groups = this.getApiGroupsForResource(rule.resource);
                apiGroups = Array.isArray(groups) ? 
                  groups.flat() : 
                  [groups];
              }
              return {
                apiGroups: apiGroups,
                resources: [resource === '*' ? resource : resource],
                verbs: verbs
              };
            });
  
          if (roleRules.length > 0) {
            configs.push(yaml.dump({
              apiVersion: 'rbac.authorization.k8s.io/v1',
              kind: 'Role',
              metadata: {
                name: role.roleName,
                namespace: this.form.namespace
              },
              rules: roleRules
            }));
  
            // RoleBinding
            configs.push(yaml.dump({
              apiVersion: 'rbac.authorization.k8s.io/v1',
              kind: 'RoleBinding',
              metadata: {
                name: role.roleBindingName || `${role.roleName}-rb`,
                namespace: this.form.namespace
              },
              subjects: [{
                kind: 'ServiceAccount',
                name: this.form.name,
                namespace: this.form.namespace
              }],
              roleRef: {
                kind: 'Role',
                name: role.roleName,
                apiGroup: 'rbac.authorization.k8s.io'
              }
            }));
          }
        }
      });
  
      // ClusterRoles
      this.form.clusterRoles.forEach(clusterRole => {
        if (clusterRole.clusterRoleName && clusterRole.rules.length > 0) {
          // ClusterRole
          const clusterRoleRules = clusterRole.rules
            .filter(rule => rule.resource && rule.verbs.length > 0)
            .map(rule => {
              const resource = rule.resource;
              const verbs = rule.verbs.length === this.allVerbs.length ? ['*'] : rule.verbs;
              
              // 修改这里：确保 apiGroups 是一维数组而不是嵌套数组
              let apiGroups;
              if (this.isWildcardResource(rule.resource) || this.hasMultipleApiGroups(rule.resource)){
                // 将 rule.apiGroup 展平为一维数组
                apiGroups = Array.isArray(rule.apiGroup) ? 
                  rule.apiGroup.flat() : 
                  [rule.apiGroup];
              } else {
                // 获取资源的 API 组并确保是一维数组
                const groups = this.getApiGroupsForResource(rule.resource);
                apiGroups = Array.isArray(groups) ? 
                  groups.flat() : 
                  [groups];
              }
              
              return {
                apiGroups: apiGroups,
                resources: [resource === '*' ? resource : resource],
                verbs: verbs
              };
            });
  
          if (clusterRoleRules.length > 0) {
            configs.push(yaml.dump({
              apiVersion: 'rbac.authorization.k8s.io/v1',
              kind: 'ClusterRole',
              metadata: {
                name: clusterRole.clusterRoleName
              },
              rules: clusterRoleRules
            }));
  
            // ClusterRoleBinding
            configs.push(yaml.dump({
              apiVersion: 'rbac.authorization.k8s.io/v1',
              kind: 'ClusterRoleBinding',
              metadata: {
                name: clusterRole.clusterRoleBindingName || `${clusterRole.clusterRoleName}-crb`
              },
              subjects: [{
                kind: 'ServiceAccount',
                name: this.form.name,
                namespace: this.form.namespace
              }],
              roleRef: {
                kind: 'ClusterRole',
                name: clusterRole.clusterRoleName,
                apiGroup: 'rbac.authorization.k8s.io'
              }
            }));
          }
        }
      });
  
      // 使用 --- 分隔每个资源
      return configs.length > 0 ? configs.join('---\n') : '';
    }
  },
  watch: {
    // 监听 namespaces 变化
    namespaces: {
      immediate: true,
      handler(val) {
        this.filteredNamespaces = val || []
      }
    }
  },
  methods: {
      // 处理对话框打开事件
      handleDialogOpen() {
      if (this.editMode && this.editData) {
        // 在编辑模式下，使用传入的数据填充表单
        this.form = JSON.parse(JSON.stringify(this.editData));
      } else {
        // 在创建模式下，重置表单
        this.resetForm();
      }
      
      // 加载资源列表
      this.loadResources();
    },
     // 重置表单
     resetForm() {
      this.form = {
        name: '',
        namespace: '',
        roles: [],
        clusterRoles: []
      };
    },
    // 添加更新方法
    async handleUpdate() {
      try {
        // 处理角色规则
        const roleRules = [];
        if (this.form.roles.length > 0 && this.form.roles[0].rules.length > 0) {
          const role = this.form.roles[0]; // 取第一个角色
          roleRules.push(...role.rules.filter(rule => rule.resource && rule.verbs.length > 0).map(rule => {
            let apiGroups;
            if (this.isWildcardResource(rule.resource) || this.hasMultipleApiGroups(rule.resource)) {
              // 将 rule.apiGroup 展平为一维数组
              apiGroups = Array.isArray(rule.apiGroup) ? 
                rule.apiGroup.flat() : 
                [rule.apiGroup];
            } else {
              // 获取资源的 API 组并确保是一维数组
              const groups = this.getApiGroupsForResource(rule.resource);
              apiGroups = Array.isArray(groups) ? 
                groups.flat() : 
                [groups];
            }
            
            // 确保 resources 是字符串数组
            // 从资源字符串中提取资源名称（去掉括号部分）
            const resourceName = rule.resource.split('(')[0].trim().toLowerCase();
            
            return {
              apiGroups: apiGroups,
              resources: [resourceName], // 将单个资源放入数组中
              verbs: rule.verbs.length === this.allVerbs.length ? ['*'] : rule.verbs
            };
          }));
        }
        
        // 处理集群角色规则
        const clusterRoleRules = [];
        if (this.form.clusterRoles.length > 0 && this.form.clusterRoles[0].rules.length > 0) {
          const clusterRole = this.form.clusterRoles[0]; // 取第一个集群角色
          clusterRoleRules.push(...clusterRole.rules.filter(rule => rule.resource && rule.verbs.length > 0).map(rule => {
            let apiGroups;
            if (this.isWildcardResource(rule.resource) || this.hasMultipleApiGroups(rule.resource)) {
              // 将 rule.apiGroup 展平为一维数组
              apiGroups = Array.isArray(rule.apiGroup) ? 
                rule.apiGroup.flat() : 
                [rule.apiGroup];
            } else {
              // 获取资源的 API 组并确保是一维数组
              const groups = this.getApiGroupsForResource(rule.resource);
              apiGroups = Array.isArray(groups) ? 
                groups.flat() : 
                [groups];
            }
            
            // 修复：添加缺失的 resourceName 定义
            const resourceName = rule.resource.split('(')[0].trim().toLowerCase();
            
            return {
              apiGroups: apiGroups,
              resources: [resourceName], // 将单个资源放入数组中
              verbs: rule.verbs.length === this.allVerbs.length ? ['*'] : rule.verbs
            };
          }));
        }
        
        // 构建符合后端结构的请求参数
        const requestData = {
          serviceAccountName: this.form.name,
          namespace: this.form.namespace,
          roleName: this.form.roles.length > 0 ? this.form.roles[0].roleName : '',
          roleBindingName: this.form.roles.length > 0 ? this.form.roles[0].roleBindingName || `${this.form.roles[0].roleName}-rb` : '',
          clusterRoleName: this.form.clusterRoles.length > 0 ? this.form.clusterRoles[0].clusterRoleName : '',
          clusterRoleBindingName: this.form.clusterRoles.length > 0 ? this.form.clusterRoles[0].clusterRoleBindingName || `${this.form.clusterRoles[0].clusterRoleName}-crb` : '',
          roleRules: roleRules,
          clusterRoleRules: clusterRoleRules
        };
        
        console.log('发送到后端的数据:', JSON.stringify(requestData, null, 2));
        
        // 调用更新 API
        const response = await updateServiceAccount(requestData);
        
        if (!response.data.error) {
          this.$message.success('ServiceAccount 更新成功');
          this.showDialog = false;
          this.$emit('updated');
        } else {
          this.$message.error(response.data.message || '更新失败');
        }
      } catch (error) {
        console.error('更新 ServiceAccount 失败:', error);
        this.$message.error('更新失败，请检查网络或服务器状态');
      }
    },
    filterNamespace(query) {
      if (!this.namespaces) {
        this.filteredNamespaces = []
        return
      }
      
      if (query) {
        this.filteredNamespaces = this.namespaces.filter(ns => 
          ns.toLowerCase().includes(query.toLowerCase())
        )
      } else {
        this.filteredNamespaces = this.namespaces
      }
    },
    addRole() {
      const roleName = '';
      this.form.roles.push({
        roleName: roleName,
        roleBindingName: `${roleName}-rb`,
        nameSpace: this.form.namespace,
        rules: []
      })
    },
    
    updateRoleBindingName(role) {
      if (!role.roleBindingNameCustomized) {
        role.roleBindingName = role.roleName ? `${role.roleName}-rb` : '';
      }
    },
    
    addClusterRole() {
      const clusterRoleName = '';
      this.form.clusterRoles.push({
        clusterRoleName: clusterRoleName,
        clusterRoleBindingName: `${clusterRoleName}-crb`,
        nameSpace: this.form.namespace,
        rules: []
      })
    },
    
    updateClusterRoleBindingName(clusterRole) {
      if (!clusterRole.clusterRoleBindingNameCustomized) {
        clusterRole.clusterRoleBindingName = clusterRole.clusterRoleName ? `${clusterRole.clusterRoleName}-crb` : '';
      }
    },
    removeRole(index) {
      this.form.roles.splice(index, 1)
    },
    addRule(role) {
      role.rules.push({
        resource: '',
        apiGroup: [],
        verbs: []
      })
    },
    removeRule(role, index) {
      role.rules.splice(index, 1);
    },
    addClusterRole() {
      this.form.clusterRoles.push({
        clusterRoleName: '',
        nameSpace: this.form.namespace,
        rules: []
      })
    },
    removeClusterRole(index) {
      this.form.clusterRoles.splice(index, 1)
    },
    addClusterRule(clusterRole) {
      clusterRole.rules.push({
        resource: '',
        apiGroup: [],
        verbs: []
      })
    },
    removeClusterRule(clusterRole, index) {
      clusterRole.rules.splice(index, 1)
    },
    async handleCreate() {
      try {
        // 处理角色规则
        const roleRules = [];
        if (this.form.roles.length > 0 && this.form.roles[0].rules.length > 0) {
          const role = this.form.roles[0]; // 取第一个角色
          roleRules.push(...role.rules.filter(rule => rule.resource && rule.verbs.length > 0).map(rule => {
            let apiGroups;
              if (this.isWildcardResource(rule.resource) || this.hasMultipleApiGroups(rule.resource)) {
                // 将 rule.apiGroup 展平为一维数组
                apiGroups = Array.isArray(rule.apiGroup) ? 
                  rule.apiGroup.flat() : 
                  [rule.apiGroup];
              } else {
                // 获取资源的 API 组并确保是一维数组
                const groups = this.getApiGroupsForResource(rule.resource);
                apiGroups = Array.isArray(groups) ? 
                  groups.flat() : 
                  [groups];
              }
            
            // 确保 resources 是字符串数组
            // 从资源字符串中提取资源名称（去掉括号部分）
            const resourceName = rule.resource.split('(')[0].trim().toLowerCase();
            
            return {
              apiGroups: apiGroups,
              resources: [resourceName], // 将单个资源放入数组中
              verbs: rule.verbs.length === this.allVerbs.length ? ['*'] : rule.verbs
            };
          }));
        }
        
        // 处理集群角色规则
        const clusterRoleRules = [];
        if (this.form.clusterRoles.length > 0 && this.form.clusterRoles[0].rules.length > 0) {
          const clusterRole = this.form.clusterRoles[0]; // 取第一个集群角色
          clusterRoleRules.push(...clusterRole.rules.filter(rule => rule.resource && rule.verbs.length > 0).map(rule => {
            let apiGroups;
              if (this.isWildcardResource(rule.resource) || this.hasMultipleApiGroups(rule.resource)) {
                // 将 rule.apiGroup 展平为一维数组
                apiGroups = Array.isArray(rule.apiGroup) ? 
                  rule.apiGroup.flat() : 
                  [rule.apiGroup];
              } else {
                // 获取资源的 API 组并确保是一维数组
                const groups = this.getApiGroupsForResource(rule.resource);
                apiGroups = Array.isArray(groups) ? 
                  groups.flat() : 
                  [groups];
              }
            
            // 修复：添加缺失的 resourceName 定义
            const resourceName = rule.resource.split('(')[0].trim().toLowerCase();
            
            return {
              apiGroups: apiGroups,
              resources: [resourceName], // 将单个资源放入数组中
              verbs: rule.verbs.length === this.allVerbs.length ? ['*'] : rule.verbs
            };
          }));
        }
        
        // 构建符合后端结构的请求参数
        const requestData = {
          serviceAccountName: this.form.name,
          namespace: this.form.namespace,
          roleName: this.form.roles.length > 0 ? this.form.roles[0].roleName : '',
          roleBindingName: this.form.roles.length > 0 ? this.form.roles[0].roleBindingName || `${this.form.roles[0].roleName}-rb` : '',
          clusterRoleName: this.form.clusterRoles.length > 0 ? this.form.clusterRoles[0].clusterRoleName : '',
          clusterRoleBindingName: this.form.clusterRoles.length > 0 ? this.form.clusterRoles[0].clusterRoleBindingName || `${this.form.clusterRoles[0].clusterRoleName}-crb` : '',
          roleRules: roleRules,
          clusterRoleRules: clusterRoleRules
        };
        
        console.log('发送到后端的数据:', JSON.stringify(requestData, null, 2));
        
        // 调用 API
        const response = await createServiceAccount(requestData);
        
        if (response.data.errorCode==0) {
          this.$message.success('ServiceAccount 创建成功');
          this.showDialog = false;
          this.$emit('created');
        } else {
          this.$message.error(response.data.errorMessage || '创建失败');
        }
      } catch (error) {
        console.error('创建 ServiceAccount 失败:', error);
        this.$message.error('创建失败，请检查网络或服务器状态');
      }
    },

    // 加载资源
    async loadResources() {
      if (!this.allResources || this.allResources.length === 0) {
        try {
          const response = await getResources()
          this.allResources = Array.isArray(response.data?.resources) ? 
            response.data.resources.map(res => ({
              Kind: res.name,
              Namespaced: res.namespaced,
              APIGroup: res.apiGroup || []
            })) : []
          // 保存 API Groups 列表
          this.apiGroupsList = response.data?.apiGroups || []
        } catch (error) {
          console.error('加载资源失败:', error)
          this.$message.error('加载资源失败，请稍后重试')
          this.allResources = []
        }
      }
    },
    // 过滤资源
    filteredResources(flag) {
      if (!Array.isArray(this.allResources)) {
        return [];
      }
      if (flag) {
        return this.allResources.filter(resource => {
        // 确保资源存在且 Namespaced 为 true
        return resource && resource.Namespaced === true;
      });
      }
      return this.allResources
     
    },
    getApiGroupsForResource(resourceValue) {
      if (!resourceValue) return [];
      // 从资源字符串中提取资源名称（去掉括号部分）
      const resourceName = resourceValue.split('(')[0];
      // 查找匹配的资源
      const resource = this.allResources.find(r => {
        const rName = r.Kind.split('(')[0];
        return rName === resourceName;
      });
      return resource && resource.APIGroup ? [resource.APIGroup] : [''];
    },
    forceUpdate() {
      this.$forceUpdate();
    },
    // 删除这里的 computed 对象
    // 新增：解析 YAML 并填充表单
    parseYaml() {
      if (!this.importYaml.trim()) {
        this.$message.warning('请先输入 YAML 内容');
        return;
      }

      try {
        // 分割多个 YAML 文档
        const yamlDocs = this.importYaml.split('---').filter(doc => doc.trim());
        
        // 重置表单
        this.form = {
          name: '',
          namespace: '',
          roles: [],
          clusterRoles: []
        };
        
        // 解析每个 YAML 文档
        yamlDocs.forEach(doc => {
          const resource = yaml.load(doc);
          
          if (!resource || !resource.kind) {
            return;
          }
          
          // 处理 ServiceAccount
          if (resource.kind === 'ServiceAccount') {
            this.form.name = resource.metadata?.name || '';
            this.form.namespace = resource.metadata?.namespace || '';
          }
          
          // 处理 Role
          else if (resource.kind === 'Role') {
            const role = {
              roleName: resource.metadata?.name || '',
              roleBindingName: resource.metadata?.name + '-rb',
              nameSpace: resource.metadata?.namespace || this.form.namespace,
              rules: []
            };
            
            // 处理规则
            if (Array.isArray(resource.rules)) {
              resource.rules.forEach(rule => {
                if (rule.resources && rule.verbs) {
                  rule.resources.forEach(resourceName => {
                    // 查找匹配的资源
                    const matchedResource = this.findResourceByName(resourceName, rule.apiGroups?.[0] || '');
                    
                    // 处理 verbs 为 * 的情况
                    const verbs = rule.verbs.includes('*') ? this.allVerbs : rule.verbs;
                    
                    role.rules.push({
                      resource: matchedResource || `${resourceName}`,
                      apiGroup: rule.apiGroups || [''],
                      verbs: verbs
                    });
                  });
                }
              });
            }
            
            this.form.roles.push(role);
          }
          
          // 处理 ClusterRole
          else if (resource.kind === 'ClusterRole') {
            const clusterRole = {
              clusterRoleName: resource.metadata?.name || '',
              clusterRoleBindingName: resource.metadata?.name + '-crb',
              nameSpace: this.form.namespace,
              rules: []
            };
            
            // 处理规则
            if (Array.isArray(resource.rules)) {
              resource.rules.forEach(rule => {
                if (rule.resources && rule.verbs) {
                  rule.resources.forEach(resourceName => {
                    // 查找匹配的资源
                    const matchedResource = this.findResourceByName(resourceName, rule.apiGroups?.[0] || '');
                    
                    // 处理 verbs 为 * 的情况
                    const verbs = rule.verbs.includes('*') ? this.allVerbs : rule.verbs;
                    
                    clusterRole.rules.push({
                      resource: matchedResource || `${resourceName}`,
                      apiGroup: rule.apiGroups || [''],
                      verbs: verbs
                    });
                  });
                }
              });
            }
            
            this.form.clusterRoles.push(clusterRole);
          }
        });
        this.$message.success('YAML 解析成功');
        // 预加载资源列表
        this.loadResources();
      } catch (error) {
        console.error('解析 YAML 失败:', error);
        this.$message.error('解析 YAML 失败，请检查格式是否正确');
      }
    },
    
    // 新增：根据资源名称和 API 组查找资源
    findResourceByName(resourceName, apiGroup) {
      if (!Array.isArray(this.allResources) || this.allResources.length === 0) {
        return null;
      }
      
      const resource = this.allResources.find(r => {
        const name = r.Kind.split('-')[0];
        return name === resourceName && r.APIGroup === apiGroup;
      });
      
      return resource ? resource.Kind : null;
    },
    
    isWildcardResource(resource) {
      return resource === '*';
    },

    hasMultipleApiGroups(resourceValue) {
      if (!resourceValue) return false;
      const resource = this.allResources.find(r => r.Kind === resourceValue);
      return resource && Array.isArray(resource.APIGroup) && resource.APIGroup.length > 1;
    },

    getAvailableApiGroups(resourceValue) {
      if (this.isWildcardResource(resourceValue)) {
        return this.apiGroupsList || [];
      }
      const resource = this.allResources.find(r => r.Kind === resourceValue);
      return resource && Array.isArray(resource.APIGroup) ? resource.APIGroup : [''];
    },

    handleResourceChange(resourceValue, rule) {
      if (!resourceValue) {
        rule.apiGroup = [];
        return;
      }
      
      // 如果是通配符或有多个 API Group，初始化为空数组
      if (this.isWildcardResource(resourceValue) || this.hasMultipleApiGroups(resourceValue)) {
        rule.apiGroup = [];
      } else {
        const resource = this.allResources.find(r => r.Kind === resourceValue);
        rule.apiGroup = resource && resource.APIGroup ? [resource.APIGroup] : [''];
      }
      
      this.forceUpdate();
    },
  }
}
</script>
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

.rule-row:last-child {
  border-bottom: none;
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

.el-checkbox {
  margin: 0;
  min-width: 80px;
}

/* 将 yaml-preview 样式移到这里 */
.yaml-preview {
  background-color: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  font-family: monospace;
  white-space: pre-wrap;
  font-size: 14px;
  line-height: 1.5;
  max-height: 400px;
  overflow-y: auto;
}

.api-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.api-tag {
  margin-right: 8px;
}

.placeholder {
  color: #909399;
  font-size: 14px;
}

.api-group-select {
  .el-select__tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  
  .el-tag {
    margin-right: 0;
    background-color: #f0f9eb;
    border-color: #e1f3d8;
    color: #67c23a;
  }
}

.el-select-dropdown.api-group-select {
  .el-select-dropdown__item.selected {
    color: #67c23a;
    font-weight: bold;
  }
}
</style>
