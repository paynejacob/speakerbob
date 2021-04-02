import { Sound } from '@/definitions/sound'

declare module 'vue/types/vue' {
  interface Vue {
    $audioPlayer: {
      EnqueueSound: (sound: Sound) => Promise<any>;
      ForceEnableSound: () => {};
    };
  }
}
