<template>
  <div>
    <v-form ref="form" v-model="valid" @submit="save" onSubmit="return false;" v-if="!!!token.id">
      <v-container>
        <v-row>
          <v-col>
            <v-text-field v-model="name" :rules="nameRules" label="Name" />
          </v-col>
        </v-row>
        <v-row>
          <v-btn block color="primary" @click="save">Save</v-btn>
        </v-row>
      </v-container>
    </v-form>
    <v-container v-if="!!token.id">
      <v-row>
          <p class="body-1">Your new API Token has been created.  Please save it somewhere secure, this value will only be shown once.</p>
      </v-row>
      <v-row>
        <v-text-field :value="token.token" outlined readonly class="centered-input" success />
      </v-row>
      <v-row>
        <v-btn block color="primary" @click="close">Close</v-btn>
      </v-row>
    </v-container>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Token } from '@/definitions/token'

@Component
export default class CreateAPIToken extends Vue {
  private valid = false;

  private name = '';
  private nameRules: any[] = [
    (v: any) => !!v || 'Name is required'
  ];

  private token: Token = new Token();

  private async save () {
    const form: any = this.$refs.form

    if (!form.validate()) {
      return
    }

    const resp: any = await this.$auth.post('/tokens/', {
      name: this.name
    })

    this.token = resp.data
  }

  private close () {
    this.reset()
    this.$emit('close')
  }

  public reset () {
    this.name = ''
    this.token = new Token()
  }
}

</script>

<style  lang="scss" scoped>
.centered-input >>> input {
  text-align: center
}
</style>
