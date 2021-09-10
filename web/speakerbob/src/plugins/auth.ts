import { Vue as _Vue } from 'vue/types/vue'
import axios, { AxiosRequestConfig } from 'axios'

export class APIOptions {}

function validateStatus (status: number): boolean {
  // any 2xx response is valid
  if (status >= 200 && status <= 299) {
    return true
  }

  // if we get an auth error send the user to the login page
  if (status === 401) {
    window.location.href = '/login'
  }

  // it is better to let the caller decide if this is valid or not
  return status === 404
}

const axiosConfig: AxiosRequestConfig = {
  baseURL: '/auth/',
  validateStatus,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
}

export default class API {
  public static install (Vue: typeof _Vue, _options?: APIOptions) {
    Vue.prototype.$auth = axios.create(axiosConfig)
  }
}
