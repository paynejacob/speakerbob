<template>
  <v-menu v-if=!disabled offset-y>
    <template v-slot:activator="{ on }">
      {{ user.display_name }}
      <v-icon v-on="on">fa-caret-down</v-icon>
    </template>
    <v-list>
      <v-spacer />
      <v-list-item @click="gotoLogout">
        <v-list-item-title>Logout</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'
import { User } from '@/definitions/user'
import router from '@/router'
import axios from 'axios'

@Component
export default class UserMenu extends Vue {
  public disabled = false
  public user: User = new User();

  public async created () {
    const resp = await axios.get('/auth/user/')

    if (resp.status === 404) {
      this.disabled = true
      return
    }

    this.user = resp.data
  }

  private gotoLogout () {
    router.push('logout')
  }
}
</script>
