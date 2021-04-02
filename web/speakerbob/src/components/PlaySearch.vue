<template>
  <v-list flat>
    <v-subheader><v-text-field prepend-icon="fa-search" v-model="query" /></v-subheader>
    <v-subheader v-if="sounds.length > 0">Sounds</v-subheader>
    <v-list-item-group>
      <v-list-item v-for="(sound, i) in sounds" :key="i" @click="playSound(sound.id)">
        <v-list-item-icon>
          <v-icon>fa-volume-up</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title v-text="sound.name"></v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list-item-group>
    <v-subheader v-if="groups.length > 0">Groups</v-subheader>
    <v-list-item-group>
      <v-list-item v-for="(group, i) in groups" :key="i" @click="playGroup(group.id)">
        <v-list-item-icon>
          <v-icon>fa-layer-group</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title v-text="group.name"></v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list-item-group>
  </v-list>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import axios from 'axios'
import { Sound } from '@/definitions/sound'
import { Group } from '@/definitions/group'

@Component
export default class PlaySearch extends Vue {
  private query = '';
  private sounds: Sound[] = [];
  private groups: Group[] = [];
  private timerId = 0

  created () {
    this.$ws.RegisterMessageHook('sound.sound.create', this.refresh)
    this.$ws.RegisterMessageHook('sound.group.create', this.refresh)
  }

  destroyed () {
    this.$ws.DeRegisterMessageHook('sound.sound.create', this.refresh)
    this.$ws.DeRegisterMessageHook('sound.group.create', this.refresh)
  }

  mounted () {
    this.search('')
  }

  public async refresh () {
    this.query = ''

    await this.search(this.query)
  }

  @Watch('query')
  private async search (query: string) {
    clearTimeout(this.timerId)

    this.timerId = setTimeout(async () => {
      const resp = await axios.request({
        method: 'get',
        url: '/sound/',
        params: {
          q: query
        }
      })

      if (resp.data) {
        this.sounds = resp.data.sounds
        this.groups = resp.data.groups
      } else {
        this.sounds = []
        this.groups = []
      }
    }, 250)
  }

  private async playSound (soundId: string) {
    await axios.request({
      method: 'PUT',
      url: `/play/sound/${soundId}/`
    })
  }

  private async playGroup (groupId: string) {
    await axios.request({
      method: 'PUT',
      url: `/play/group/${groupId}/`
    })
  }
}
</script>

<style scoped>

</style>
