<template>
  <v-app>
    <v-app-bar app dark color="primary">
      <div class="d-flex align-center">
        <v-app-bar-title>Speakerbob</v-app-bar-title>
      </div>
      <v-spacer/>
      <ConnectionStatus :connected="connected" />
      <UserCount :user-count="userCount" />
    </v-app-bar>
    <v-main>
      <v-card id="create" height="100%">
        <v-container fluid>
          <v-row>
            <v-col offset-md="3" md="6" sm="12">
              <PlaySearch />
            </v-col>
          </v-row>
        </v-container>
        <v-speed-dial top right absolute direction="bottom">
          <template v-slot:activator>
            <v-btn v-model="fab" color="blue darken-2" dark fab>
              <v-icon v-if="fab">
                fa-close
              </v-icon>
              <v-icon v-else>
                fa-plus
              </v-icon>
            </v-btn>
          </template>
          <v-btn fab dark small color="green" @click="createSoundModal = !createSoundModal">
            <v-icon>fa-volume-up</v-icon>
          </v-btn>
        </v-speed-dial>
        <v-dialog v-model="createSoundModal">
          <v-card>
            <v-card-title>Create Sound</v-card-title>
            <v-card-text>
              <CreateSound ref="createSoundForm" @submit="() => createSoundModal = false" />
            </v-card-text>
          </v-card>
        </v-dialog>
      </v-card>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue'
import PlaySearch from '@/components/PlaySearch.vue'
import CreateSound from '@/components/CreateSound.vue'
import { Component, Watch } from 'vue-property-decorator'
import UserCount from '@/components/UserCount.vue'
import ConnectionStatus from '@/components/ConnectionStatus.vue'

@Component({ components: { ConnectionStatus, UserCount, PlaySearch, CreateSound } })
export default class App extends Vue {
  private fab = false;
  private createSoundModal = false;
  private connected = false;
  private userCount = 0;
  private connection!: WebSocket;

  created () {
    this.connect()
  }

  private connect () {
    const proto = (window.location.protocol === 'https:') ? 'wss' : 'ws'

    this.connection = new WebSocket(`${proto}://${window.location.hostname}:${window.location.port}/ws/`)

    this.connection.onopen = this.connectionOpen
    this.connection.onclose = this.connectionClose

    this.connection.onmessage = this.readMessage
  }

  private connectionOpen () {
    this.connected = true
  }

  private connectionClose () {
    this.connected = false

    setTimeout(this.connect, 500)
  }

  private readMessage (event: MessageEvent) {
    const message = JSON.parse(event.data)

    switch (message.type) {
      case 'connection_count':
        this.userCount = message.payload.count
        break
      case 'play':
        const audio = new Audio(`/sound/${message.payload.sound.id}/download/`)
        audio.play()
    }
  }

  private resetCreateSoundForm () {
    const createSoundForm: any = this.$refs.createSoundForm
    createSoundForm.reset()
  }

  @Watch('createSoundModal')
  private createSoundModalChange (value: boolean) {
    if (value) {
      return
    }

    this.resetCreateSoundForm()
  }
}
</script>
