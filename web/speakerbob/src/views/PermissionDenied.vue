<template>
  <v-container fill-height fluid>
    <v-row align="center" justify="center">
      <v-col md="3">
        <v-card>
          <v-card-title>Permission Denied</v-card-title>
          <v-card-text>The user you have signed in with does not have permission to access this application.</v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component } from 'vue-property-decorator'

@Component({})
export default class PermissionDenied extends Vue {}
</script>
