import { RouteConfig } from 'vue-router'
import Home from '@/views/Home.vue'

const routes: Array<RouteConfig> = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/login/',
    name: 'Login',
    meta: { disableWS: true },
    component: () => import('@/views/Login.vue')
  },
  {
    path: '/logout/',
    name: 'Logout',
    meta: { disableWS: true },
    component: () => import('@/views/Logout.vue')
  },
  {
    path: '/userpreferences/',
    name: 'UserPreferences',
    meta: { disableWS: false },
    component: () => import('@/views/UserPreferences.vue')
  },
  {
    path: '/permission-denied/',
    name: 'PermissionDenied',
    meta: { disableWS: true },
    component: () => import('@/views/PermissionDenied.vue')
  },
  {
    path: '*',
    redirect: { name: 'Home' }
  }
]

export default routes
