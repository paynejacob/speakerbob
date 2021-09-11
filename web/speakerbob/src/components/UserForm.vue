<template>
  <v-form ref="form" v-model="valid" @submit="() => {}" onSubmit="return false;" :readonly="loading">
    <v-container fluid>
      <v-row>
        <v-col>
          <v-text-field v-model="name" label="Name" :loading="loading" />
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <SoundSelect v-model="entrySound" label="Entrance Sound" :defaultSound="entrySound" />
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { Sound } from '@/definitions/sound'
import SoundSelect from '@/components/SoundSelect.vue'
import { UserPreferences } from '@/definitions/userpreferences'
@Component({
  components: { SoundSelect }
})
export default class UserForm extends Vue {
  private timerId = 0
  private loading = true;

  private valid = false;

  private name = '';
  private entrySound: Sound = {};

  public async created () {
    const preferences: UserPreferences = (await this.$auth.get('/user/preferences/')).data
    this.name = preferences.name || ''

    if (preferences.entrySoundId) {
      const sound: Sound = (await this.$api.get(`/sound/sounds/${preferences.entrySoundId}/`)).data
      this.entrySound = sound
    }

    this.loading = false
  }

  @Watch('name')
  private async onNameChange (name: string) {
    await this.save({ name: name })
  }

  @Watch('entrySound')
  private async onSoundChange (sound: Sound) {
    await this.save({ entrySoundId: sound.id })
  }

  private async save (preferences: UserPreferences) {
    clearTimeout(this.timerId)

    this.timerId = setTimeout(async () => {
      this.loading = true

      await this.$auth.patch('/user/preferences/', preferences)

      this.loading = false
    }, 500)
  }
}

</script>
