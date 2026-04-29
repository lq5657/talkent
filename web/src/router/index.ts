import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'setup',
      component: () => import('../views/SetupView.vue'),
    },
    {
      path: '/chat/:id',
      name: 'chat',
      component: () => import('../views/ChatView.vue'),
    },
    {
      path: '/report/:id',
      name: 'report',
      component: () => import('../views/ReportView.vue'),
    },
  ],
})

export default router
