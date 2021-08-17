<template>
  <div v-if=!connected>
    <v-icon class="red--text">fa-wifi</v-icon>
  </div>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component } from 'vue-property-decorator'

@Component
export default class ConnectionStatus extends Vue {
  private connected = false;

  public created () {
    this.$ws.RegisterConnectionHook(this.setConnected)
  }

  public destroyed () {
    this.$ws.DeRegisterConnectionHook(this.setConnected)
  }

  private async setConnected (connected: boolean) {
    this.connected = connected
  }
}
</script>
