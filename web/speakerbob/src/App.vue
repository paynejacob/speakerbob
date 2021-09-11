<template>
  <v-app>
    <v-app-bar app dark color="primary">
      <div class="d-flex align-center">
          <v-toolbar-title @click="$router.push('/')">
            Speakerbob
          </v-toolbar-title>
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
import { UserPreferences } from '@/definitions/userpreferences'
import Cookies from 'js-cookie'

@Component({ components: { ConnectionStatus, UserCount, UserMenu } })
export default class App extends Vue {
  private showOverlay = false;

  private wsEnabled = true;

  public async created () {
    // start listening for play requests
    this.$ws.RegisterMessageHook('play', this.onPlayMessage)
    this.$ws.RegisterConnectionHook(this.playEntrySound)
    this.$ws.Connect()
  }

  public destroyed () {
    this.$ws.DeRegisterMessageHook('play', this.onPlayMessage)
  }

  @Watch('$route')
  public toggleWS (to: Route) {
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

  private async playEntrySound (connected: boolean) {
    const skipCookieName = 'skipEntrySound'

    // check for a skip cookie
    if (Cookies.get(skipCookieName)) {
      return
    }

    // load the user's entry sound
    const preferences: UserPreferences = (await this.$auth.get('/user/preferences/')).data

    // if the user does not have an entry sound exit
    if (!preferences.entrySoundId) {
      return
    }

    // play the user's entry sound
    await this.$api.put(`/sound/sounds/${preferences.entrySoundId}/play/`)

    // set the skip cookie
    const expires = new Date()
    expires.setMinutes(expires.getMinutes() + 15)
    Cookies.set(skipCookieName, 'true', { expires, path: '/noop' })
  }
}
</script>
