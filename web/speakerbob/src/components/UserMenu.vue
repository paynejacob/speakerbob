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
      <v-list-item @click="logout">
        <v-list-item-title>Logout</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import { UserPreferences } from '@/definitions/userpreferences'
import axios from 'axios'

@Component
export default class UserMenu extends Vue {
  public disabled = false
  public user: UserPreferences = new UserPreferences();

  public async created () {
    try {
      const resp = await axios.get('/auth/user/preferences/')
      this.user = resp.data
    } catch (e) {
      if (axios.isAxiosError(e)) {
        if (!!e.response && (e.response.status === 404 || e.response.status === 401)) {
          this.disabled = true
          return
        }
      }
      throw e
    }
  }

  private async goto (path: string) {
    await this.$router.push(path)
  }

  private logout () {
    this.disabled = true
    this.goto('logout')
  }
}
</script>
