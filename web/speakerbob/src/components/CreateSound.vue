<template>
  <v-form ref="form" v-model="valid">
    <v-container>
      <v-row>
        <v-col>
          <v-text-field v-model="name" :rules="nameRules" label="Name" />
        </v-col>
      </v-row>
      <v-row>
        <v-col>
          <v-file-input v-model="file" :rules="fileRules" label="sound file" />
        </v-col>
        <v-col md="2">
          <v-switch v-model="nsfw" label="NSFW" />
        </v-col>
      </v-row>
      <v-row>
        <v-btn block color="primary" @click="saveSound">Save</v-btn>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import {Component, Vue, Watch} from 'vue-property-decorator'
import axios from 'axios'
import { Sound } from '@/definitions/sound'

@Component
export default class CreateSound extends Vue {
  private valid: boolean = false;

  private file: any = null;
  private fileRules: any[] = [
    (v: any) => !!v || 'Sound file is required',
  ];

  private name = '';
  private nameRules: any[] = [
    (v: any) => !!v || 'Name is required'
  ];

  public nsfw: boolean = false;

  private sound: Sound = {
    id: '',
    name: '',
    nsfw: false
  };

  @Watch('file')
  private async uploadFile (file: File | null) {
    if (file === null) return

    const form = new FormData()
    form.append(file.name, file)

    const resp = await axios.request({
      method: 'POST',
      url: '/sound/',
      data: form,
      headers: {
        'content-type': 'multipart/form-data'
      }
    })

    this.sound = resp.data
  }

  private async saveSound () {
    const form: any = this.$refs.form

    this.sound.name = this.name
    this.sound.nsfw = this.nsfw

    if (!form.validate()) {
      return
    }

    await axios.request({
      method: 'PATCH',
      url: `/sound/${this.sound.id}/`,
      data: this.sound
    })

    this.reset()

    this.$emit('submit')
  }

  public reset() {
    const form: any = this.$refs.form
    form.reset()
  }
}

</script>
