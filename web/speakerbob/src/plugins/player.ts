import { Sound } from '@/definitions/sound'
import { Vue as _Vue } from 'vue/types/vue'
import { WebsocketOptions } from '@/plugins/websocket'
import { AxiosInstance, AxiosResponse } from 'axios'

class AsyncBlockingQueue {
  private promises: Promise<Sound>[] = [];
  private resolvers: ((t: Sound) => void)[] = [];

  private add () {
    this.promises.push(new Promise(resolve => {
      this.resolvers.push(resolve)
    }))
  }

  public enqueue (sound: Sound) {
    if (!this.resolvers.length) this.add()
    const resolve = this.resolvers.shift()!
    resolve(sound)
  }

  public dequeue (): Promise<Sound> {
    if (!this.promises.length) this.add()
    return this.promises.shift()!
  }

  public isEmpty (): boolean {
    return !this.promises.length
  }
}

export default class Player {
  private enabled = false
  private isPlaying = false
  private queue: AsyncBlockingQueue = new AsyncBlockingQueue()

  private ctx!: AudioContext
  private api!: AxiosInstance

  public constructor (api: AxiosInstance) {
    this.install = this.install.bind(this)
    this.OnPlayMessage = this.OnPlayMessage.bind(this)
    this.EnableSound = this.EnableSound.bind(this)
    this.playNextSound = this.playNextSound.bind(this)

    this.api = api
  }

  public install (Vue: typeof _Vue, _options?: WebsocketOptions) {
    Vue.prototype.$player = this
  }

  public async OnPlayMessage (message: any) {
    if (!this.enabled) {
      return
    }

    // add this sound to the queue
    this.queue.enqueue(message.sound)

    if (!this.isPlaying) {
      await this.playNextSound()
    }
  }

  public async EnableSound () {
    this.ctx = this.ctx = new window.AudioContext()

    this.enabled = true
  }

  private async playNextSound () {
    this.isPlaying = true

    const sound = await this.queue.dequeue()

    const resp: AxiosResponse = await this.api.get(`/sound/sounds/${sound.id}/download/`, { responseType: 'arraybuffer' })

    const buf = await this.ctx.decodeAudioData(resp.data)
    const source = this.ctx.createBufferSource()
    source.buffer = buf
    source.connect(this.ctx.destination)
    source.start()
    source.onended = this.playNextSound
  }
}
