<template>
  <v-form ref="form" v-model="valid">
    <v-container fluid>
      <v-row>
        <v-col>
          <v-text-field v-model="name" :rules="nameRules" label="Name" />
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-stepper vertical>
            <v-stepper-items>
              <template v-for="(sound, i) in sounds">
                <v-stepper-step :complete="true" :key="i" :step="i">{{sound.name}}<small class="remove-text" @click="() => removeSound(i)">remove</small></v-stepper-step>
                <v-stepper-content :key="i" :step="i" />
              </template>
            </v-stepper-items>
            <v-stepper-content :step="sounds.length + 1">
              <v-autocomplete auto-select-first
                              placeholder="add sound..."
                              prepend-icon="fa-plus"
                              :items="searchResults"
                              :loading="loading"
                              v-model="searchModel"
                              :search-input.sync="query"
                              item-text="name"
                              return-object
                              :rules="soundRules"/>
            </v-stepper-content>
          </v-stepper>
        </v-col>
      </v-row>
      <v-row>
        <v-btn block color="primary" @click="save">Save</v-btn>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import axios from 'axios'
import { Sound } from '@/definitions/sound'

@Component
export default class CreateGroup extends Vue {
  private valid = false;
  private sounds: Sound[] = [];

  private timerId = 0
  private loading = false;
  private searchResults: Sound[] = [];
  private query = '';
  private searchModel: any = null;
  private soundRules: any[] = [
    () => { if (this.sounds.length < 2) { return 'at least 2 sounds are required' } return true }
  ];

  private name = '';
  private nameRules: any[] = [
    (v: any) => !!v || 'Name is required'
  ];

  public nsfw = false;

  private async save () {
    const form: any = this.$refs.form

    if (!form.validate()) {
      return
    }

    await axios.request({
      method: 'POST',
      url: '/sound/group/',
      data: {
        name: this.name,
        sounds: this.sounds.map((s: Sound) => s.id)
      }
    })

    this.reset()

    this.$emit('submit')
  }

  public reset () {
    const form: any = this.$refs.form
    form.reset()
    this.searchModel = null
    this.query = ''
    this.searchResults = []
    this.sounds = []
  }

  @Watch('query')
  private async search (query: string) {
    clearTimeout(this.timerId)

    this.timerId = setTimeout(async () => {
      this.loading = true

      if (!query) {
        this.searchResults = []
        this.loading = false
        return
      }

      const resp = await axios.request({
        method: 'get',
        url: '/sound/',
        params: {
          q: query
        }
      })

      if (resp.data) {
        this.searchResults = resp.data.sounds
      } else {
        this.searchResults = []
      }
      this.loading = false
    }, 250)
  }

  @Watch('searchModel')
  private addSound (sound: Sound | null) {
    if (sound === null) {
      return
    }

    this.sounds.push(sound)
    this.searchModel = null
    this.query = ''
    this.searchResults = []
  }

  private removeSound (index: number) {
    if (index > -1) {
      this.sounds.splice(index, 1)
    }
  }
}

</script>

<style lang="scss" scoped>
.remove-text {
  &:hover {
    text-decoration: underline;
  }
}
</style>
