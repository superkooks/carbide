<script setup lang="ts">
import { useEphemeralStore } from "@/stores/ephemeral"
import { useGuildsStore } from "@/stores/guilds"

const store = useGuildsStore()
const ephem = useEphemeralStore()

// function changeName(id: string) {
//   ephem.ws?.send(
//     encrypt(
//       {
//         method: "SET",
//         path: ".name",
//         object: "wat",
//       },
//       id,
//       store.txSessions[id]
//     )
//   )
// }
// @click="changeName(guild.id)"
</script>

<template>
  <div class="container">
    <div v-for="guild in store.guilds" :key="guild.id">
      {{ guild.name }}
    </div>
  </div>
</template>

<style scoped>
.container {
  background-color: var(--md-surfaceneg1);
  height: 100%;
  width: 70px;
}
</style>
