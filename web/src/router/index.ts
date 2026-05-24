import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
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

router.beforeEach((to, _from, next) => {
  if (to.name === 'login') {
    if (useAuth().isAuthenticated()) {
      next({ path: '/' })
    } else {
      next()
    }
    return
  }
  if (!useAuth().isAuthenticated()) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
