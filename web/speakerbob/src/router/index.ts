import Vue from 'vue'
import VueRouter, { RouteConfig } from 'vue-router'
import Home from '@/views/Home.vue'

Vue.use(VueRouter)

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

export default new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})
