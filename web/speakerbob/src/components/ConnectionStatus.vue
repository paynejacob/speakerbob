<template>
  <div :hidden="connected">
    <v-icon class="red--text">fa-wifi</v-icon>
  </div>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'

@Component
export default class ConnectionStatus extends Vue {
  @Prop(Boolean) readonly connected!: boolean;
}
</script>
