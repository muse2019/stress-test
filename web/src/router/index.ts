import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      children: [
        { path: '', redirect: '/tasks' },
        { path: 'tasks', component: () => import('@/views/TaskList.vue') },
        { path: 'monitor/:id', component: () => import('@/views/Monitor.vue') },
        { path: 'reports', component: () => import('@/views/Reports.vue') },
      ]
    }
  ]
})

export default router
