import { defineStore } from "pinia"
import { utob, btou, deepCopy } from "../util";

export interface Guild {
  id: string
  name: string
  channels: Channel[]
}

export interface Channel {
  id: string
  name: string
}

export interface Mutation {
  method: "SET" | "DELETE"
  path: string
  object?: any
}

export const useGlobalStore = defineStore({
  id: "global",

  state: () => ({
    guilds: {} as {[guild: string]: Guild},
    txSessions: {} as {[guild: string]: TxSession},
  }),

  actions: {
    overwrite(state: any) {
      // Import our crypto sessions before assigning to state
      for (const guild in state.txSessions) {
        state.txSessions[guild] = window.tungsten.importTx(btou(state.txSessions[guild]))
      }

      for (const prop in state) {
        // typescript doing it's best to make my code ugly
        (this as {[key: string]: any})[prop] = state[prop]
      }

      this.guilds = {}
      this.guilds["6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b"] = {
        id: "6ec0bd7f-11c0-43da-975e-2a8ad9ebae0b",
        name: "hih",
        channels: []
      }
    },

    export(): string {
      let copy = deepCopy(this.$state) as any
      
      // Modify it to export our crypto sessions properly
      copy.txSessions = {}
      for (const guild in this.txSessions) {
        // It's fine to store tx sessions as base64 as they are relatively
        // small and are being stored locally
        copy.txSessions[guild] = utob(this.txSessions[guild].export())
      }

      return JSON.stringify(copy)
    }
  },
})
