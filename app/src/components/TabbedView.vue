<script setup lang="ts">
import { reactive } from "vue"

const props = defineProps<{ tabs: string[]; active: boolean }>()

const state = reactive({
  activeIndex: 0,
})
</script>

<template>
  <div class="view-container">
    <div class="tabs">
      <div
        class="tab body-large"
        v-for="(tab, n) in props.tabs"
        :key="n"
        :class="{ active: n == state.activeIndex }"
      >
        {{ tab }}
      </div>
    </div>

    <template v-for="(tab, n) in props.tabs">
      <div class="content body-medium" v-if="n == state.activeIndex" :key="n">
        {{ tab }}
      </div>
    </template>
  </div>
</template>

<style scoped>
.view-container {
  flex-grow: 1;
  height: 100%;

  display: flex;
  flex-direction: column;
}

.tabs {
  background-color: var(--md-surfaceneg1);
  display: flex;
  flex-direction: row;
}

.tab {
  background-color: var(--md-surface);
  padding: 10px;
  width: fit-content;
}

.tab.active {
  background: var(--md-surface1);
}
</style>
