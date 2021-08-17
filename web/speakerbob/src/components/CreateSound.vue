<template>
  <v-form ref="form" v-model="valid" @submit="saveSound" onSubmit="return false;">
    <v-container>
      <v-row>
        <v-col>
          <v-text-field v-model="name" :rules="nameRules" label="Name" />
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-file-input v-model="file" :rules="fileRules" :error="!!fileErrors.length" :error-messages="fileErrors" label="sound file" />
        </v-col>
      </v-row>
      <v-row>
        <v-btn block color="primary" @click="saveSound" :disabled="soundId === ''">Save</v-btn>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'

@Component
export default class CreateSound extends Vue {
  private valid = false;

  private file: any = null;
  private fileErrors: string[] = [];
  private fileRules: any[] = [
    (v: any) => !!v || 'Sound file is required'
  ];

  private name = '';
  private nameRules: any[] = [
    (v: any) => !!v || 'Name is required'
  ];

  private soundId = '';

  @Watch('file')
  private async uploadFile (file: File | null) {
    if (file === null) return

    const form = new FormData()
    form.append(file.name, file)

    try {
      const resp = await this.$api.post('/sounds/', form, { headers: { 'content-type': 'multipart/form-data' } })

      this.soundId = resp.data.id
      this.fileErrors = []
    } catch {
      this.fileErrors = ['invalid audio file']
    }
  }

  private async saveSound () {
    const form: any = this.$refs.form

    if (!form.validate() || !!this.fileErrors.length) {
      return
    }

    await this.$api.patch(`/api/sound/${this.soundId}/`, {
      name: this.name
    })

    this.reset()

    this.$emit('submit')
  }

  public reset () {
    const form: any = this.$refs.form
    form.reset()
    this.soundId = ''
    this.fileErrors = []
  }
}

</script>
