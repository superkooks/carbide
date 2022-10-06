import { defineStore } from "pinia"
import { utob, btou, deepCopy } from "../util";

export const useUserStore = defineStore({
  id: "user",

  state: () => ({
    id: "",
    token: null as Uint8Array|null,
    deviceTx: null as TxSession|null,
  }),

  actions: {
    overwrite(state: any) {
      // Import things properly before assigning to state
      if (state.deviceTx != undefined) { 
        state.deviceTx = window.tungsten.importTx(btou(state.deviceTx))
      }
      if (state.token != undefined) {
        state.token = btou(state.token)
      }

      for (const prop in state) {
        // typescript doing it's best to make my code ugly
        (this as {[key: string]: any})[prop] = state[prop]
      }
    },

    export(): string {
      let copy = deepCopy(this.$state) as any
      
      // Modify it to export things properly
      if (this.deviceTx != null) {
        copy.deviceTx = utob(this.deviceTx.export())
      }
      if (this.token != null) {
        copy.token = utob(this.token)
      }

      return JSON.stringify(copy)
    }
  },
})
