import { createRouter, createWebHistory } from 'vue-router'

import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { useBattleStore } from '@/stores/battle'

declare module 'vue-router' {
  interface RouteMeta {
    /** 需要登录态，未登录重定向 /login */
    requiresAuth?: boolean
    /** 仅未登录可见（登录页），已登录跳首页 */
    guestOnly?: boolean
    /** 顶栏页面标题 */
    title?: string
  }
}

const router = createRouter({
  // 生产由 Go 的 NoRoute 回落 index.html，深链接可直达；路由路径不能带 /static 前缀
  history: createWebHistory('/'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guestOnly: true, title: '登录注册' },
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'home', component: () => import('@/views/HomeView.vue'), meta: { title: '大厅' } },
        {
          path: 'stages',
          name: 'stages',
          component: () => import('@/views/StagesView.vue'),
          meta: { title: '西行关卡' },
        },
        {
          path: 'heroes',
          name: 'heroes',
          component: () => import('@/views/HeroesView.vue'),
          meta: { title: '英雄阵容' },
        },
        {
          path: 'gacha',
          name: 'gacha',
          component: () => import('@/views/GachaView.vue'),
          meta: { title: '天命召唤' },
        },
        {
          path: 'star-up',
          name: 'starUp',
          component: () => import('@/views/StarUpView.vue'),
          meta: { title: '碎片升星' },
        },
        {
          path: 'activity',
          name: 'activity',
          component: () => import('@/views/ActivityView.vue'),
          meta: { title: '限时活动' },
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: { name: 'home' } },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'home' }
  }
  // 战斗锁导航：战斗未结束时禁止离开关卡页（替代旧 iframe postMessage 锁）
  const battle = useBattleStore()
  if (battle.isLocked && to.name !== 'stages') {
    useToast().show('战斗进行中，请先完成战斗或投降')
    return false
  }
  return true
})

export default router
