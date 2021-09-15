<template>
  <v-app>
    <v-app-bar app dark color="primary" @click="$router.push('/')">
      <div class="d-flex align-center">
          <v-toolbar-title>
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
    </v-main>
    <v-overlay :value="showOverlay" opacity="1" color="primary" @click.native="dismissOverlay">
      <h1>Click to Start Speakerbob</h1>
    </v-overlay>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component, Watch } from 'vue-property-decorator'
import ConnectionStatus from '@/components/ConnectionStatus.vue'
import UserCount from '@/components/UserCount.vue'
import UserMenu from '@/components/UserMenu.vue'
import { Route } from 'vue-router'
import Cookies from 'js-cookie'
import { UserPreferences } from '@/definitions/userpreferences'

@Component({ components: { ConnectionStatus, UserCount, UserMenu } })
export default class App extends Vue {
  private showOverlay = true;
  private wsEnabled = true;

  @Watch('$route')
  public toggleWS (to: Route) {
    if (!to.meta) {
      return
    }

    this.wsEnabled = !to.meta.disableWS
    this.showOverlay = this.wsEnabled && this.showOverlay
  }

  private async dismissOverlay () {
    await this.$player.EnableSound()
    this.showOverlay = false
    await this.playEntrySound()
  }

  private async playEntrySound () {
    const timeoutKey = 'entrySoundTimeout'

    // check timeout
    const timeoutRaw = localStorage.getItem(timeoutKey)
    if (timeoutRaw) {
      const now: Date = new Date()
      const timeout: Date = new Date(parseInt(timeoutRaw))

      if (now < timeout) {
        return
      }
    }

    // load the user's entry sound
    const preferences: UserPreferences = (await this.$auth.get('/user/preferences/')).data

    // if the user does not have an entry sound exit
    if (!preferences.entrySoundId) {
      return
    }

    // play the user's entry sound
    await this.$api.put(`/sound/sounds/${preferences.entrySoundId}/play/`)

    // set timeout
    const newTimeout = new Date()
    newTimeout.setMinutes(newTimeout.getMinutes() + 15)
    localStorage.setItem(timeoutKey, newTimeout.getTime().toString())
  }
}
</script>
