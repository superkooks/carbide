import { defineStore } from "pinia"

export const useEphemeralStore = defineStore({
  id: "ephemeral",

  state: () => ({
    ws: null as WebSocket|null
  }),
})
