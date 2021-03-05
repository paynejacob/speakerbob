<template>
  <v-list flat>
    <v-subheader><v-text-field prepend-icon="fa-search" v-model="query" /></v-subheader>
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
  </v-list>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import axios from 'axios'
import { Sound } from '@/definitions/sound'

@Component
export default class PlaySearch extends Vue {
  private query = '';
  private isLoading = false;
  private sounds: Sound[] = [];
  private timerId = 0

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
        this.sounds = resp.data
      } else {
        this.sounds = []
      }
    }, 250)
  }

  private async playSound (soundId: string) {
    await axios.request({
      method: 'PUT',
      url: `/play/sound/${soundId}/`
    })

    this.query = ''
  }
}
</script>

<style scoped>

</style>
