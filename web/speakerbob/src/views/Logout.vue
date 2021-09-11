<template>
  <v-container fill-height fluid>
    <v-row align="center" justify="center">
      <v-col md="3">
        <v-card v-if="loggedOut">
          <v-card-title class="justify-center">You are now logged out</v-card-title>
          <v-card-actions>
            <v-btn block color="primary" @click=gotoLogin>Login</v-btn>
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
export default class Logout extends Vue {
  private loggedOut = false;

  public async mounted () {
    this.loggedOut = false

    await axios.get('/auth/logout/')

    this.loggedOut = true
  }

  private async gotoLogin () {
    await this.$router.push('/login')
  }
}
</script>
