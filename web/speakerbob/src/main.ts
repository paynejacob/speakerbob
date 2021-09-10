import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import Player from './plugins/player'
import '@babel/polyfill'
import 'roboto-fontface/css/roboto/roboto-fontface.css'
import '@fortawesome/fontawesome-free/css/all.css'
import WSConnection from '@/plugins/websocket'
import wb from './registerServiceWorker'
import router from './router'
import API from '@/plugins/api'
import Auth from '@/plugins/auth'

Vue.config.productionTip = false

Vue.use(Player)
Vue.use(WSConnection)
Vue.use(API)
Vue.use(Auth)

Vue.prototype.$workbox = wb

new Vue({
  vuetify,
  router,
  render: h => h(App)
}).$mount('#app')
