<template>
  <v-app>
    <v-app-bar app dark color="primary">
      <div class="d-flex align-center">
        <v-app-bar-title>Speakerbob</v-app-bar-title>
      </div>
      <v-spacer/>
      <v-app-bar-nav-icon v-if="wsEnabled">
        <ConnectionStatus />
      </v-app-bar-nav-icon>
      <v-app-bar-nav-icon v-if="wsEnabled">
        <UserCount />
      </v-app-bar-nav-icon>
      <v-app-bar-nav-icon>
        <UserMenu />
      </v-app-bar-nav-icon>
    </v-app-bar>
    <v-main>
      <router-view />
      <v-overlay :value="showOverlay">
        <v-btn @click="dismissOverlay">Click here to start Speakerbob</v-btn>
      </v-overlay>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component, Watch } from 'vue-property-decorator'
import ConnectionStatus from '@/components/ConnectionStatus.vue'
import UserCount from '@/components/UserCount.vue'
import UserMenu from '@/components/UserMenu.vue'
import { Route } from 'vue-router'

@Component({ components: { ConnectionStatus, UserCount, UserMenu } })
export default class App extends Vue {
  private showOverlay = false;

  private wsEnabled = true;

  public async created () {
    // start listening for play requests
    this.$ws.RegisterMessageHook('play', this.onPlayMessage)
    this.$ws.Connect()
  }

  public destroyed () {
    this.$ws.DeRegisterMessageHook('play', this.onPlayMessage)
  }

  @Watch('$route')
  public toggleWS (to: Route, from: Route) {
    if (to.meta !== undefined) {
      this.wsEnabled = !to.meta.disableWS
    } else {
      this.wsEnabled = true
    }

    if (this.wsEnabled) {
      this.$ws.Connect()
    } else {
      this.$ws.Stop()
    }
  }

  private dismissOverlay () {
    this.$audioPlayer.ForceEnableSound()

    this.showOverlay = false
  }

  private async onPlayMessage (message: any) {
    try {
      await this.$audioPlayer.EnqueueSound(message.sound)
    } catch (e) {
      if (e.name === 'NotAllowedError') {
        this.showOverlay = true
      }
    }
  }
}
</script>
