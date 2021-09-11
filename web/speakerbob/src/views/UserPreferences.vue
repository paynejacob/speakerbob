<template>
  <v-container fluid>
    <v-tabs v-model="tab">
      <v-tab>Profile</v-tab>
      <v-tab>API Tokens</v-tab>
    </v-tabs>
    <v-tabs-items v-model="tab">
      <v-tab-item>
        <UserForm />
      </v-tab-item>
      <v-tab-item>
        <APITokenTable />
      </v-tab-item>
    </v-tabs-items>
  </v-container>
</template>

<script lang="ts">
import Vue from 'vue'
import { Component } from 'vue-property-decorator'
import APITokenTable from '@/components/APITokenTable.vue'
import UserForm from '@/components/UserForm.vue'

@Component({
  components: { UserForm, APITokenTable }
})
export default class UserPreferences extends Vue {
  private tab = 0;
}
</script>
