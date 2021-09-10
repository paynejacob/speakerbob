<template>
  <v-card>
    <v-data-table
      :headers="headers"
      :items="tokens"
      disable-pagination
      hide-default-footer>
      <template v-slot:top>
        <v-toolbar flat>
          <v-spacer />
          <v-btn color="primary" @click="() => createTokenModal = !createTokenModal">Create</v-btn>
        </v-toolbar>
      </template>
      <template v-slot:item.actions="{ item }">
        <v-icon
          small
          @click="deleteToken(item)">
          fa-trash
        </v-icon>
      </template>
    </v-data-table>
    <v-dialog v-model="createTokenModal">
      <v-card>
        <v-card-title>Create API Token</v-card-title>
        <v-card-text>
          <CreateAPIToken ref="createAPITokenForm" @close="() => createTokenModal = false"/>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component, Watch } from 'vue-property-decorator'
import { Token } from '@/definitions/token'
import CreateAPIToken from '@/components/CreateAPIToken.vue'

@Component({
  components: { CreateAPIToken }
})
export default class APITokenTable extends Vue {
  private readonly headers: any[] = [
    { text: 'Name', value: 'name', align: 'start' },
    { text: 'ID', value: 'id' },
    { text: 'Created At', value: 'created_at' },
    { text: 'Actions', value: 'actions', sortable: false }
  ];

  private createTokenModal = false;

  private tokens: Token[] = [];

  $refs!: {
    createAPITokenForm: HTMLFormElement;
  }

  public async created () {
    await this.getTokens()
  }

  @Watch('createTokenModal')
  private async resetCreateGroupForm (value: boolean) {
    if (!value) {
      this.$refs.createAPITokenForm.reset()
    }
    await this.getTokens()
  }

  private async getTokens () {
    const resp = await this.$auth.get('/tokens/')
    this.tokens = resp.data
  }

  private async deleteToken (token: Token) {
    await this.$auth.delete(`/tokens/${token.id}/`)
    await this.getTokens()
  }
}
</script>
