import { Vue as _Vue } from 'vue/types/vue'
import wb from '@/registerServiceWorker'
export class APIOptions {}

export default class Workbox {
  public install (Vue: typeof _Vue, _options?: APIOptions) {
    Vue.prototype.$wb = wb
  }
}
