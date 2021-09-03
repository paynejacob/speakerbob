import { Sound } from '@/definitions/sound'
import { Vue as _Vue } from 'vue/types/vue'

export class PlayerOptions {}

export default class Player {
  private audio!: HTMLAudioElement;
  private queue: Sound[] = [];
  private isPlaying = false;

  public constructor () {
    this.audio = new Audio()
  }

  public async EnqueueSound (sound: Sound) {
    this.queue.push(sound)

    if (!this.isPlaying) {
      await this.playNextSound()
    }
  }

  public ForceEnableSound () {
    this.audio.src = ''
    this.audio.play()
  }

  public static install (Vue: typeof _Vue, _options?: PlayerOptions) {
    Vue.prototype.$audioPlayer = new Player()
  }

  private async playNextSound () {
    while (true) {
      const sound = this.queue.pop()

      if (sound === undefined) {
        return
      }

      this.audio.src = `/sound/sounds/${sound.id}/download/`

      this.isPlaying = true

      await this.audio.play()

      this.isPlaying = false
    }
  }
}
