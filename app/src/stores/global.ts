import { defineStore } from "pinia"

export const useGlobalStore = defineStore({
  id: "global",

  state: () => ({
    counter: 0,
  }),

  getters: {
    doubleCount: (state) => state.counter * 2,
  },

  actions: {
    overwrite(state: any) {
      this.counter = state.counter
    },

    increment() {
      this.counter++
    },
  },
})
