<template>
  <v-autocomplete auto-select-first
                  placeholder="add sound..."
                  :items="searchResults"
                  :loading="loading"
                  v-model="sound"
                  :search-input.sync="query"
                  item-text="name"
                  :label="label"
                  return-object />
</template>

<script lang="ts">
import { Component, Prop, VModel, Vue, Watch } from 'vue-property-decorator'
import { Sound } from '@/definitions/sound'

@Component
export default class SoundSelect extends Vue {
  @VModel() sound!: Sound
  @Prop() readonly defaultSound!: Sound;
  @Prop() readonly label!: string;

  private timerId = 0
  private loading = false;
  private searchResults: Sound[] = [];
  private query = '';

  @Watch('query')
  private async search (query: string) {
    if (query === this.defaultSound.name) {
      return
    }

    clearTimeout(this.timerId)

    this.timerId = setTimeout(async () => {
      this.loading = true

      if (!query) {
        this.searchResults = []
        this.loading = false
        return
      }

      const resp = await this.$api.get(`/sound/search/?q=${escape(query)}`)

      if (resp.data) {
        this.searchResults = resp.data.sounds
      } else {
        this.searchResults = []
      }
      this.loading = false
    }, 250)
  }

  @Watch('defaultSound')
  private onDefaultQueryChange (sound: Sound) {
    this.searchResults = [sound]
  }
}

</script>
