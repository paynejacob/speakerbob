<template>
  <v-container fill-height fluid>
    <v-row align="center" justify="center">
      <v-col md="3">
        <v-card>
          <v-card-title>Login</v-card-title>
          <v-card-actions>
            <v-btn block v-for="provider in providers" @click="() => loginWithProvider(provider)" :key="provider">{{ provider }}</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component } from 'vue-property-decorator'
import axios from 'axios'

@Component({})
export default class Login extends Vue {
  private providers: string[] = [];

  public async created () {
    this.providers = await Login.getProviders()

    if (this.providers.length === 0) {
      return await this.$router.push('/')
    }
  }

  private static async getProviders (): Promise<string[]> {
    const resp = await axios.get('/auth/providers/')

    return resp.data
  }

  private loginWithProvider (provider: string) {
    window.location.href = `/auth/login/?provider=${provider}`
  }
}
</script>
