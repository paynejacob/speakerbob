import { Vue as _Vue } from 'vue/types/vue'
import VueRouter, { NavigationGuardNext, Route } from 'vue-router'
import { AxiosInstance } from 'axios'
import { Token } from '@/definitions/token'

export class WebsocketOptions {
  router!: VueRouter
}

export type MessageHookFn = {(message: any): Promise<void> | void}
export type ConnectionHookFn = {(connected: boolean): Promise<void>}

export default class WSConnection {
  private connection!: WebSocket;
  private auth!: AxiosInstance

  private messageHooks!: Map<string, MessageHookFn[]>
  private connectionHooks: ConnectionHookFn[] = []

  private stopped = true
  private connected = false
  private connecting = false

  constructor (authAPI: AxiosInstance) {
    this.install = this.install.bind(this)
    this.RegisterMessageHook = this.RegisterMessageHook.bind(this)
    this.DeRegisterMessageHook = this.DeRegisterMessageHook.bind(this)
    this.RegisterConnectionHook = this.RegisterConnectionHook.bind(this)
    this.DeRegisterConnectionHook = this.DeRegisterConnectionHook.bind(this)
    this.Connect = this.Connect.bind(this)
    this.NavigationGuard = this.NavigationGuard.bind(this)
    this.connectionOpen = this.connectionOpen.bind(this)
    this.connectionClose = this.connectionClose.bind(this)
    this.readMessage = this.readMessage.bind(this)
    this.handleVisibilityChange = this.handleVisibilityChange.bind(this)

    this.auth = authAPI
    this.messageHooks = new Map<string, MessageHookFn[]>()

    // iOS Safari can freeze/kill a backgrounded tab's socket without firing
    // 'close', leaving no reconnect scheduled once the tab foregrounds again.
    document.addEventListener('visibilitychange', this.handleVisibilityChange)
    window.addEventListener('pageshow', this.handleVisibilityChange)
  }

  public install (Vue: typeof _Vue, _options?: WebsocketOptions) {
    Vue.prototype.$ws = this
  }

  public RegisterMessageHook (type: string, hook: MessageHookFn) {
    const hooks = this.messageHooks.get(type) || []

    hooks.push(hook)
    this.messageHooks.set(type, hooks)
  }

  public DeRegisterMessageHook (type: string, hook: MessageHookFn) {
    const hooks = this.messageHooks.get(type) || []

    const index = hooks.indexOf(hook)
    if (index > -1) {
      hooks.splice(index, 1)
      this.messageHooks.set(type, hooks)
    }
  }

  public RegisterConnectionHook (hook: ConnectionHookFn) {
    this.connectionHooks.push(hook)
  }

  public DeRegisterConnectionHook (hook: ConnectionHookFn) {
    const index = this.connectionHooks.indexOf(hook)
    if (index > -1) {
      this.connectionHooks.splice(index, 1)
    }
  }

  public async Connect () {
    if (this.connected || this.connecting) {
      return
    }

    this.connecting = true

    try {
      const token: Token = (await this.auth.get('/tokens/ws/')).data
      const proto = (window.location.protocol === 'https:') ? 'wss' : 'ws'
      const port = window.location.port ? `:${window.location.port}` : ''
      const url = `${proto}://${window.location.hostname}${port}/ws/?token=${token.token}`

      this.stopped = false

      this.connection = new WebSocket(url)

      this.connection.onopen = () => this.connectionOpen()
      this.connection.onclose = () => this.connectionClose()
      this.connection.onmessage = (props) => this.readMessage(props)
    } catch (error) {
      this.connecting = false
      throw error
    }
  }

  public Stop () {
    this.stopped = true

    if (this.connection) {
      this.connection.close()
    }
  }

  public NavigationGuard (to: Route, _: Route, next: NavigationGuardNext) {
    // if the route explicitly disable ws stop the ws
    if (!!to.meta && to.meta.disableWS) {
      this.Stop()
    } else {
      this.Connect()
    }

    next()
  }

  private async connectionOpen () {
    this.connected = true
    this.connecting = false

    for (let i = 0; i < this.connectionHooks.length; i++) {
      await this.connectionHooks[i](true)
    }
  }

  private async connectionClose () {
    this.connected = false
    this.connecting = false

    for (let i = 0; i < this.connectionHooks.length; i++) {
      await this.connectionHooks[i](false)
    }

    if (!this.stopped) {
      setTimeout(() => this.Connect(), Math.random() * 1000)
    }
  }

  private handleVisibilityChange () {
    if (document.visibilityState !== 'visible' || this.stopped) {
      return
    }

    const readyState = this.connection?.readyState
    if (readyState === undefined || readyState === WebSocket.CLOSED || readyState === WebSocket.CLOSING) {
      this.Connect()
    }
  }

  private async readMessage (event: MessageEvent) {
    const message: any = JSON.parse(event.data)

    const hooks = this.messageHooks.get(message.type)
    if (hooks !== undefined) {
      for (let i = 0; i < hooks.length; i++) {
        await hooks[i](message)
      }
    }
  }
}
