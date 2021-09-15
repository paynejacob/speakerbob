import { Vue as _Vue } from 'vue/types/vue'
import axios, { AxiosInstance } from 'axios'
import VueRouter from 'vue-router'
export class APIOptions {}

export default class API {
  private router!: VueRouter
  public readonly api!: AxiosInstance

  constructor (router: VueRouter) {
    this.router = router

    this.validateStatus = this.validateStatus.bind(this)

    this.api = axios.create({
      baseURL: '/api/',
      validateStatus: this.validateStatus,
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json'
      }
    })
  }

  public install (Vue: typeof _Vue, _options?: APIOptions) {
    Vue.prototype.$api = this.api
  }

  private validateStatus (status: number): boolean {
    // any 2xx response is valid
    if (status >= 200 && status <= 299) {
      return true
    }

    // if we get an auth error send the user to the login page
    if (status === 401) {
      this.router.push({ name: 'Login' })
    }

    // it is better to let the caller decide if this is valid or not
    return status === 404
  }
}
