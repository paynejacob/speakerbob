<template>
  <v-tooltip bottom>
    <template v-slot:activator="{ on }">
      <div v-on="on">
        <v-icon>fa-user</v-icon>
        {{ userCount }}
      </div>
    </template>
    <span v-if="onlineUsers.length">{{ onlineUsers.join(', ') }}</span>
    <span v-else>No one signed in</span>
  </v-tooltip>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'

@Component
export default class UserCount extends Vue {
  private userCount = 0
  private onlineUsers: string[] = []

  public created () {
    this.$ws.RegisterMessageHook('connection_count', this.setUserCount)
    this.$ws.RegisterMessageHook('presence', this.setOnlineUsers)
  }

  public destroyed () {
    this.$ws.DeRegisterMessageHook('connection_count', this.setUserCount)
    this.$ws.DeRegisterMessageHook('presence', this.setOnlineUsers)
  }

  private async setUserCount (message: any) {
    this.userCount = message.count
  }

  private async setOnlineUsers (message: any) {
    this.onlineUsers = message.users
  }
}
</script>
