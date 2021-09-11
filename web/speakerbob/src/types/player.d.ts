import { Sound } from '@/definitions/sound'

declare module 'vue/types/vue' {
  interface Vue {
    $player: {
      EnqueueSound(sound: Sound): Promise<any>;
      EnableSound(): void;
      OnPlayMessage(message: any): Promise<never>;
    };
  }
}
