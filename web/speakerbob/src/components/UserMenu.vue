<template>
  <v-menu v-if=!disabled offset-y>
    <template v-slot:activator="{ on }">
      <v-icon v-on="on">fa-caret-down</v-icon>
    </template>
    <v-list>
      <v-list-item>{{user.name}}</v-list-item>
      <v-divider />
      <v-list-item @click="goto('userpreferences')">
        <v-list-item-title>Preferences</v-list-item-title>
      </v-list-item>
      <v-spacer />
      <v-list-item @click="goto('logout')">
        <v-list-item-title>Logout</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import { UserPreferences } from '@/definitions/userpreferences'

@Component
export default class UserMenu extends Vue {
  public disabled = false
  public user: UserPreferences = new UserPreferences();

  public async created () {
    const resp = await this.$auth.get('/user/preferences/')

    if (resp.status === 404) {
      this.disabled = true
      return
    }

    this.user = resp.data
  }

  private async goto (path: string) {
    await this.$router.push(path)
  }
}
</script>
