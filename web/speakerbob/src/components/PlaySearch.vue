<template>
  <v-list flat>
    <v-subheader><v-text-field prepend-icon="fa-search" v-model="query" /></v-subheader>
    <v-subheader>
      <v-select :items="sortOptions" v-model="sortSelection" label="Sort by" dense hide-details />
    </v-subheader>
    <v-subheader v-if="sounds.length > 0">Sounds</v-subheader>
    <v-list-item-group>
      <v-list-item v-for="(sound, i) in sounds" :key="i" @click="playSound(sound.id)">
        <v-list-item-icon>
          <v-icon>fa-volume-up</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title v-text="sound.name"></v-list-item-title>
          <v-list-item-subtitle>{{ sound.play_count || 0 }} plays</v-list-item-subtitle>
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
import { Sound } from '@/definitions/sound'
import { Group } from '@/definitions/group'

@Component
export default class PlaySearch extends Vue {
  private query = '';
  private sounds: Sound[] = [];
  private groups: Group[] = [];
  private timerId = 0
  private sortSelection = ''
  private readonly sortOptions = [
    { text: 'Relevance', value: '' },
    { text: 'Name (A-Z)', value: 'name:asc' },
    { text: 'Most played', value: 'play_count:desc' },
    { text: 'Newest', value: 'created_at:desc' }
  ]

  created () {
    this.$ws.RegisterMessageHook('update_sound', this.onUpdateSound)
    this.$ws.RegisterMessageHook('delete_sound', this.onDeleteSound)

    this.$ws.RegisterMessageHook('create_group', this.onCreateGroup)
    this.$ws.RegisterMessageHook('update_group', this.onUpdateGroup)
    this.$ws.RegisterMessageHook('delete_group', this.onDeleteGroup)
  }

  destroyed () {
    this.$ws.DeRegisterMessageHook('update_sound', this.onUpdateSound)
    this.$ws.DeRegisterMessageHook('delete_sound', this.onDeleteSound)

    this.$ws.DeRegisterMessageHook('create_group', this.onCreateGroup)
    this.$ws.DeRegisterMessageHook('update_group', this.onUpdateGroup)
    this.$ws.DeRegisterMessageHook('delete_group', this.onDeleteGroup)
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
      const params = new URLSearchParams({ q: query })
      const [sortField, sortOrder] = this.sortSelection.split(':')
      if (sortField) {
        params.set('sort', sortField)
      }
      if (sortOrder) {
        params.set('order', sortOrder)
      }

      const resp = await this.$api.get(`/sound/search/?${params.toString()}`)

      if (resp.data) {
        this.sounds = resp.data.sounds
        this.groups = resp.data.groups
      } else {
        this.sounds = []
        this.groups = []
      }
    }, 250)
  }

  @Watch('sortSelection')
  private onSortSelectionChange () {
    this.search(this.query)
  }

  private async playSound (soundId: string) {
    await this.$api.put(`/sound/sounds/${soundId}/play/`)
  }

  private async playGroup (groupId: string) {
    await this.$api.put(`/sound/groups/${groupId}/play/`)
  }

  private onUpdateSound (message: any) {
    const sound = message.sound

    for (let i = 0; i < this.sounds.length - 1; i++) {
      if (this.sounds[i].id === sound.id) {
        this.sounds[i] = sound
        return
      }
    }

    this.sounds = [sound].concat(this.sounds)
  }

  private onDeleteSound (message: any) {
    const sound = message.sound

    for (let i = this.sounds.length - 1; i >= 0; i--) {
      if (this.sounds[i].id === sound.id) {
        this.sounds.splice(i, 1)
      }
    }
  }

  private onCreateGroup (message: any) {
    const group = message.group

    this.groups = [group].concat(this.groups)
  }

  private onUpdateGroup (message: any) {
    const group = message.group

    for (let i = 0; i < this.groups.length - 1; i++) {
      if (this.groups[i].id === group.id) {
        this.groups[i] = group
        return
      }
    }

    this.groups = [group].concat(this.groups)
  }

  private onDeleteGroup (message: any) {
    const group = message.group

    for (let i = this.groups.length - 1; i >= 0; i--) {
      if (this.groups[i].id === group.id) {
        this.groups.splice(i, 1)
      }
    }
  }
}
</script>

<style scoped>

</style>
