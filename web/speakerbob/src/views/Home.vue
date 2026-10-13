<template>
  <v-card id="create" class="home-screen">
    <v-container fluid class="home-screen__body">
      <v-row class="home-screen__row flex-nowrap">
        <v-col class="home-screen__list d-flex flex-column" offset-md="3" md="6" cols="8">
          <PlaySearch />
        </v-col>
        <v-col class="home-screen__actions d-flex justify-center" md="3" cols="4">
          <v-speed-dial class="d-flex justify-center" direction="bottom">
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
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<style scoped>
.home-screen {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  height: calc(100dvh - 64px);
  overflow: hidden;
}

.home-screen__body {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.home-screen__row {
  flex: 1 1 auto;
  min-height: 0;
}

.home-screen__list {
  min-width: 0;
  min-height: 0;
  height: 100%;
}

.home-screen__actions {
  align-self: flex-start;
  position: relative;
  z-index: 2;
}
</style>

<script lang="ts">
import Vue from 'vue'
import { Component, Watch } from 'vue-property-decorator'
import PlaySearch from '@/components/PlaySearch.vue'
import ConnectionStatus from '@/components/ConnectionStatus.vue'
import UserCount from '@/components/UserCount.vue'

const CreateSound = () => import('@/components/CreateSound.vue')
const CreateGroup = () => import('@/components/CreateGroup.vue')
const Say = () => import('@/components/Say.vue')

@Component({ components: { CreateGroup, ConnectionStatus, UserCount, PlaySearch, CreateSound, Say } })
export default class Home extends Vue {
  private fab = false;
  private createSoundModal = false;
  private createGroupModal = false;
  private sayModal = false;

  $refs!: {
    createSoundForm: HTMLFormElement;
    createGroupForm: HTMLFormElement;
    sayForm: HTMLFormElement;
  }

  @Watch('createSoundModal')
  private resetCreateSoundForm (value: boolean) {
    if (!value) {
      this.$refs.createSoundForm.reset()
    }
  }

  @Watch('createGroupModal')
  private resetCreateGroupForm (value: boolean) {
    if (!value) {
      this.$refs.createGroupForm.reset()
    }
  }

  @Watch('sayModal')
  private resetSayForm (value: boolean) {
    if (!value) {
      this.$refs.sayForm.reset()
    }
  }
}
</script>
