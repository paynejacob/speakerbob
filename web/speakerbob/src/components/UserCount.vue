<template>
  <div>
    <v-icon>fa-user</v-icon>
    {{ userCount }}
  </div>
</template>

<script lang="ts">
import { Vue, Component, Prop } from 'vue-property-decorator'

@Component
export default class UserCount extends Vue {
  @Prop(Number) readonly userCount!: number;
}
</script>

<style scoped>

</style>
