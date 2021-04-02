import { ConnectionHookFn, MessageHookFn } from '@/plugins/websocket'

declare module 'vue/types/vue' {
  interface Vue {
    $ws: {
      RegisterMessageHook(type: string, hook: MessageHookFn): void;
      DeRegisterMessageHook(type: string, hook: MessageHookFn): void;
      RegisterConnectionHook(hook: ConnectionHookFn): void;
      DeRegisterConnectionHook(hook: ConnectionHookFn): void;
    };
  }
}
