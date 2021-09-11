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
  private wsEnabled = true;

  @Watch('$route')
  public toggleWS (to: Route) {
    if (!to.meta) {
      return
    }

    this.wsEnabled = !to.meta.disableWS
  }
}
</script>
