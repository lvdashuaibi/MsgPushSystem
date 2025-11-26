import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      }
    ]
  },
  {
    path: '/messages',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '消息管理' },
    children: [
      {
        path: '',
        name: 'Messages',
        redirect: '/messages/send'
      },
      {
        path: 'send',
        name: 'MessageSend',
        component: () => import('@/views/Messages/Send.vue'),
        meta: { title: '发送消息' }
      },
      {
        path: 'templates',
        name: 'MessageTemplates',
        component: () => import('@/views/Messages/Templates.vue'),
        meta: { title: '消息模板' }
      },
      {
        path: 'records',
        name: 'MessageRecords',
        component: () => import('@/views/Messages/Records.vue'),
        meta: { title: '消息记录' }
      }
    ]
  },
  {
    path: '/users',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '用户管理' },
    children: [
      {
        path: '',
        name: 'Users',
        redirect: '/users/list'
      },
      {
        path: 'list',
        name: 'UserList',
        component: () => import('@/views/Users/List.vue'),
        meta: { title: '用户列表' }
      },
      {
        path: 'tags',
        name: 'UserTags',
        component: () => import('@/views/Users/Tags.vue'),
        meta: { title: '标签管理' }
      }
    ]
  },
  {
    path: '/scheduled',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: '定时消息' },
    children: [
      {
        path: '',
        name: 'Scheduled',
        redirect: '/scheduled/list'
      },
      {
        path: 'list',
        name: 'ScheduledList',
        component: () => import('@/views/Scheduled/List.vue'),
        meta: { title: '定时列表' }
      },
      {
        path: 'create',
        name: 'ScheduledCreate',
        component: () => import('@/views/Scheduled/Create.vue'),
        meta: { title: '创建定时消息' }
      }
    ]
  },
  {
    path: '/ai-polish',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { title: 'AI润色' },
    children: [
      {
        path: '',
        name: 'AIPolish',
        component: () => import('@/views/AIPolish/index.vue'),
        meta: { title: 'AI内容润色' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
