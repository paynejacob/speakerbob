<template>
  <div>
    <v-icon>fa-user</v-icon>
    {{ userCount }}
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'

@Component
export default class UserCount extends Vue {
  private userCount = 0

  public created () {
    this.$ws.RegisterMessageHook('connection_count', this.setUserCount)
  }

  public destroyed () {
    this.$ws.DeRegisterMessageHook('connection_count', this.setUserCount)
  }

  private async setUserCount (message: any) {
    this.userCount = message.count
  }
}
</script>
