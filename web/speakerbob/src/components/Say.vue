<template>
  <v-form ref="form" v-model="valid" @submit="save" onSubmit="return false;">
    <v-container fluid>
      <v-row>
        <v-col>
          <v-text-field v-model="message" :rules="messageRules" placeholder="say something here..." label="Message" />
        </v-col>
      </v-row>
      <v-row>
        <v-btn block color="primary" :loading="loading" @click="save">Say</v-btn>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Say extends Vue {
  private valid = false;

  private loading = false;

  private message = '';
  private messageRules: any[] = [
    (v: any) => !!v || 'message is required'
  ];

  private async save () {
    const form: any = this.$refs.form

    if (!form.validate()) {
      return
    }

    this.loading = true

    await this.$api.put('/sound/say/', JSON.stringify(this.message))

    this.loading = false

    this.reset()

    this.$emit('submit')
  }

  public reset () {
    const form: any = this.$refs.form
    form.reset()
  }
}

</script>
