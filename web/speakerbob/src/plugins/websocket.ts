import { Vue as _Vue } from 'vue/types/vue'

export class WebsocketOptions {}

export type MessageHookFn = {(message: Message): Promise<void>}
export type ConnectionHookFn = {(connected: boolean): Promise<void>}

export interface Message {
  type: string;
  payload: any;
}

export default class WSConnection {
  private readonly url!: string;
  private connection!: WebSocket;
  private messageHooks!: Map<string, MessageHookFn[]>
  private connectionHooks: ConnectionHookFn[] = []

  constructor (url: string) {
    this.url = url
    this.messageHooks = new Map<string, MessageHookFn[]>()

    this.connect()
  }

  public static install (Vue: typeof _Vue, _options?: WebsocketOptions) {
    const proto = (window.location.protocol === 'https:') ? 'wss' : 'ws'
    Vue.prototype.$ws = new WSConnection(`${proto}://${window.location.hostname}:${window.location.port}/ws/`)
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

  private connect () {
    this.connection = new WebSocket(this.url)

    this.connection.onopen = () => this.connectionOpen()
    this.connection.onclose = () => this.connectionClose()
    this.connection.onmessage = (props) => this.readMessage(props)
  }

  private async connectionOpen () {
    for (let i = 0; i < this.connectionHooks.length; i++) {
      await this.connectionHooks[i](true)
    }
  }

  private async connectionClose () {
    for (let i = 0; i < this.connectionHooks.length; i++) {
      await this.connectionHooks[i](false)
    }

    setTimeout(() => this.connect(), Math.random() * 1000)
  }

  private async readMessage (event: MessageEvent) {
    const message: Message = JSON.parse(event.data)

    const hooks = this.messageHooks.get(message.type)
    if (hooks !== undefined) {
      for (let i = 0; i < hooks.length; i++) {
        await hooks[i](message)
      }
    }
  }
}
