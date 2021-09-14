import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import Player from './plugins/player'
import '@babel/polyfill'
import 'roboto-fontface/css/roboto/roboto-fontface.css'
import '@fortawesome/fontawesome-free/css/all.css'
import WSConnection from '@/plugins/websocket'
import API from '@/plugins/api'
import routes from '@/routes'
import VueRouter from 'vue-router'
import Auth from '@/plugins/auth'
import Workbox from '@/plugins/workbox'

Vue.config.productionTip = false

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: routes
})

const wsConnection = new WSConnection()
const player = new Player()
const api = new API(router)
const auth = new Auth(router)
const workbox = new Workbox()

router.beforeEach(wsConnection.NavigationGuard)
wsConnection.RegisterMessageHook('play', player.OnPlayMessage)

Vue.use(VueRouter)
Vue.use(wsConnection)
Vue.use(player)
Vue.use(api)
Vue.use(auth)
Vue.use(workbox)

new Vue({
  vuetify,
  router,
  render: h => h(App)
}).$mount('#app')
