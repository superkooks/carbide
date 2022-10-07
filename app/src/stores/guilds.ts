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

export const useGuildsStore = defineStore({
  id: "guilds",

  state: () => ({
    guilds: {} as {[guild: string]: Guild},
    txSessions: {} as {[guild: string]: TxSession},
    latestTs: {} as {[guild: string]: number},
  }),

  actions: {
    overwrite(state: any) {
      // Import our crypto sessions before assigning to state
      if (state.txSessions != null) {
        for (const guild in state.txSessions) {
          state.txSessions[guild] = window.tungsten.importTx(btou(state.txSessions[guild]))
        }
      }

      for (const prop in state) {
        // typescript doing it's best to make my code ugly
        (this as {[key: string]: any})[prop] = state[prop]
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
