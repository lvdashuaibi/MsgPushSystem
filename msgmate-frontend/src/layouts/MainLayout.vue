<template>
  <div class="main-layout">
    <el-container class="layout-container">
      <!-- 侧边栏 -->
      <el-aside width="250px" class="sidebar">
        <div class="logo">
          <el-icon class="logo-icon"><Message /></el-icon>
          <span class="logo-text">MsgMate</span>
        </div>

        <el-menu
          :default-active="activeMenu"
          class="sidebar-menu"
          background-color="#001529"
          text-color="#ffffff"
          active-text-color="#1890ff"
          @select="handleMenuSelect"
        >
          <el-menu-item index="/dashboard">
            <el-icon><Odometer /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>

          <el-sub-menu index="messages">
            <template #title>
              <el-icon><ChatDotRound /></el-icon>
              <span>消息管理</span>
            </template>
            <el-menu-item index="/messages/send">发送消息</el-menu-item>
            <el-menu-item index="/messages/templates">消息模板</el-menu-item>
            <el-menu-item index="/messages/records">消息记录</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="users">
            <template #title>
              <el-icon><User /></el-icon>
              <span>用户管理</span>
            </template>
            <el-menu-item index="/users/list">用户列表</el-menu-item>
            <el-menu-item index="/users/tags">标签管理</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="scheduled">
            <template #title>
              <el-icon><Timer /></el-icon>
              <span>定时消息</span>
            </template>
            <el-menu-item index="/scheduled/list">定时列表</el-menu-item>
            <el-menu-item index="/scheduled/create">创建定时消息</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-aside>

      <!-- 主内容区域 -->
      <el-main class="main-content">
        <div class="content-header">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path" :to="item.path">
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="content-body">
          <router-view />
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Message,
  Odometer,
  ChatDotRound,
  User,
  Timer
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

// 当前激活的菜单项
const activeMenu = computed(() => {
  return route.path
})

// 面包屑导航
const breadcrumbs = computed(() => {
  const path = route.path
  const breadcrumbs = [{ path: '/dashboard', title: '首页' }]

  if (path.startsWith('/messages')) {
    breadcrumbs.push({ path: '/messages', title: '消息管理' })
    if (path === '/messages/send') {
      breadcrumbs.push({ path: '/messages/send', title: '发送消息' })
    } else if (path === '/messages/templates') {
      breadcrumbs.push({ path: '/messages/templates', title: '消息模板' })
    } else if (path === '/messages/records') {
      breadcrumbs.push({ path: '/messages/records', title: '消息记录' })
    }
  } else if (path.startsWith('/users')) {
    breadcrumbs.push({ path: '/users', title: '用户管理' })
    if (path === '/users/list') {
      breadcrumbs.push({ path: '/users/list', title: '用户列表' })
    } else if (path === '/users/tags') {
      breadcrumbs.push({ path: '/users/tags', title: '标签管理' })
    }
  } else if (path.startsWith('/scheduled')) {
    breadcrumbs.push({ path: '/scheduled', title: '定时消息' })
    if (path === '/scheduled/list') {
      breadcrumbs.push({ path: '/scheduled/list', title: '定时列表' })
    } else if (path === '/scheduled/create') {
      breadcrumbs.push({ path: '/scheduled/create', title: '创建定时消息' })
    }
  }

  return breadcrumbs
})

// 菜单选择处理
const handleMenuSelect = (index: string) => {
  router.push(index)
}
</script>

<style scoped>
.main-layout {
  height: 100vh;
  overflow: hidden;
}

.layout-container {
  height: 100%;
}

.sidebar {
  background-color: #001529;
  height: 100vh;
  overflow-y: auto;
}

.logo {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  color: #ffffff;
  border-bottom: 1px solid #1f1f1f;
}

.logo-icon {
  font-size: 24px;
  margin-right: 12px;
  color: #1890ff;
}

.logo-text {
  font-size: 20px;
  font-weight: 600;
}

.sidebar-menu {
  border: none;
  height: calc(100vh - 65px);
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
}

.sidebar-menu .el-sub-menu .el-menu-item {
  height: 40px;
  line-height: 40px;
  padding-left: 50px !important;
}

.main-content {
  background-color: #f0f2f5;
  padding: 0;
  height: 100vh;
  overflow-y: auto;
}

.content-header {
  background-color: #ffffff;
  padding: 16px 24px;
  border-bottom: 1px solid #e8e8e8;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.content-body {
  padding: 0;
  min-height: calc(100vh - 65px);
}

:deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #1890ff;
  font-weight: 500;
}
</style>
