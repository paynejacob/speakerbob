<template>
  <v-app>
    <v-app-bar app dark color="primary">
      <div class="d-flex align-center">
        <v-app-bar-title>Speakerbob</v-app-bar-title>
      </div>
      <v-spacer/>
      <ConnectionStatus />
      <UserCount />
    </v-app-bar>
    <v-main>
      <v-card id="create" height="100%">
        <v-container fluid>
          <v-row>
            <v-col offset-md="3" md="6" sm="12">
              <PlaySearch ref="playSearch" />
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
          <v-btn fab dark small color="green" @click="createGroupModal = !createGroupModal">
            <v-icon>fa-layer-group</v-icon>
          </v-btn>
          <v-btn fab dark small color="green" @click="sayModal = !sayModal">
            <v-icon>fa-comment</v-icon>
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
        <v-dialog v-model="createGroupModal">
          <v-card>
            <v-card-title>Create Group</v-card-title>
            <v-card-text>
              <CreateGroup ref="createGroupForm" @submit="() => createGroupModal = false" />
            </v-card-text>
          </v-card>
        </v-dialog>
        <v-dialog v-model="sayModal">
          <v-card>
            <v-card-title>Say</v-card-title>
            <v-card-text>
              <Say ref="sayForm" />
            </v-card-text>
          </v-card>
        </v-dialog>
      </v-card>
      <v-overlay :value="showOverlay">
        <v-btn @click="dismissOverlay">Click here to start Speakerbob</v-btn>
      </v-overlay>
    </v-main>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component, Watch } from 'vue-property-decorator'
import { Message } from '@/plugins/websocket'
import PlaySearch from '@/components/PlaySearch.vue'

const ConnectionStatus = () => import('@/components/ConnectionStatus.vue')
const UserCount = () => import('@/components/UserCount.vue')
const CreateSound = () => import('@/components/CreateSound.vue')
const CreateGroup = () => import('@/components/CreateGroup.vue')
const Say = () => import('@/components/Say.vue')

@Component({ components: { CreateGroup, ConnectionStatus, UserCount, PlaySearch, CreateSound, Say } })
export default class App extends Vue {
  private fab = false;
  private createSoundModal = false;
  private createGroupModal = false;
  private sayModal = false;

  private showOverlay = false;

  $refs!: {
    createSoundForm: HTMLFormElement;
    playSearch: PlaySearch;
  }

  public created () {
    this.$ws.RegisterMessageHook('play', this.onPlayMessage)
  }

  public destroyed () {
    this.$ws.DeRegisterMessageHook('play', this.onPlayMessage)
  }

  private async onPlayMessage (message: Message) {
    try {
      await this.$audioPlayer.EnqueueSound(message.payload.sound)
    } catch (e) {
      if (e.name === 'NotAllowedError') {
        this.showOverlay = true
      }
    }
  }

  private resetCreateSoundForm () {
    this.$refs.createSoundForm.reset()
  }

  @Watch('createSoundModal')
  private createSoundModalChange (value: boolean) {
    if (value) {
      return
    }

    this.$refs.playSearch.refresh()

    this.resetCreateSoundForm()
  }

  private dismissOverlay () {
    this.$audioPlayer.ForceEnableSound()

    this.showOverlay = false
  }
}
</script>
