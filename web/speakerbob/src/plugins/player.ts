import { Sound } from '@/definitions/sound'
import { Vue as _Vue } from 'vue/types/vue'
import { WebsocketOptions } from '@/plugins/websocket'

export default class Player {
  private enabled = false
  private audio!: HTMLAudioElement;
  private queue: Sound[] = [];
  private isPlaying = false;

  public constructor () {
    this.install = this.install.bind(this)
    this.OnPlayMessage = this.OnPlayMessage.bind(this)
    this.EnableSound = this.EnableSound.bind(this)
    this.playNextSound = this.playNextSound.bind(this)

    this.audio = new Audio()
  }

  public install (Vue: typeof _Vue, _options?: WebsocketOptions) {
    Vue.prototype.$player = this
  }

  public async OnPlayMessage (sound: any) {
    if (!this.enabled) {
      return
    }

    this.queue.push(sound.sound)

    await this.playNextSound()
  }

  public async EnableSound () {
    this.audio.src = ''

    try {
      await this.audio.play()
    } catch {}

    this.enabled = true
  }

  private async playNextSound () {
    if (this.isPlaying) {
      return
    }

    while (true) {
      const sound = this.queue.pop()

      if (sound === undefined) {
        return
      }

      this.audio.src = `/api/sound/sounds/${sound.id}/download/`

      this.isPlaying = true

      await this.audio.play()

      this.isPlaying = false
    }
  }
}
