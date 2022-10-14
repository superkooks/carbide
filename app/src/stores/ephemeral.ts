import { defineStore } from "pinia"
import type { Mutation } from "./guilds"

export const useEphemeralStore = defineStore({
  id: "ephemeral",

  state: () => ({
    ws: null as WebSocket|null,
    pendingMutations: {} as {[eventId: string]: Mutation},
    selectedGuildId: "",
  }),
})
