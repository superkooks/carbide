<script setup lang="ts">
import { reactive, ref, onMounted } from "vue"
import TabbedView from "./TabbedView.vue"

const state: {
  tabs: string[][]
  highlightLeft: boolean
  highlightRight: boolean
} = reactive({
  tabs: [],
  highlightLeft: false,
  highlightRight: false,
})

function dragOver(ev: Event) {
  if (!(ev instanceof DragEvent) || ev.dataTransfer == null) {
    return
  }

  ev.preventDefault()
  ev.dataTransfer.dropEffect = "move"

  // Translate mouse position to relative to main view
  const currentRect = (ev.currentTarget as HTMLElement).getBoundingClientRect()
  const offX = ev.pageX - currentRect.left

  // Highlight appropriate drop zone
  const elWidth = (ev.currentTarget as HTMLElement).offsetWidth
  if (state.tabs.length == 0) {
    state.highlightLeft = true
    state.highlightRight = true
  } else if (state.tabs.length == 1) {
    if (offX < (3 * elWidth) / 4) {
      state.highlightLeft = true
    } else {
      state.highlightLeft = false
    }

    if (offX > elWidth / 4) {
      state.highlightRight = true
    } else {
      state.highlightRight = false
    }
  } else {
    if (offX < elWidth / 2) {
      state.highlightLeft = true
    } else {
      state.highlightLeft = false
    }

    if (offX > elWidth / 2) {
      state.highlightRight = true
    } else {
      state.highlightRight = false
    }
  }
}

function dragLeave() {
  state.highlightLeft = false
  state.highlightRight = false
}

function drop(ev: Event) {
  if (!(ev instanceof DragEvent)) {
    return
  }
  ev.preventDefault()

  // Get ID
  const data = ev.dataTransfer?.getData("text/plain")
  if (data == undefined) {
    return
  }

  // Translate mouse position to relative to main view
  const currentRect = (ev.currentTarget as HTMLElement).getBoundingClientRect()
  const offX = ev.pageX - currentRect.left

  // Add the tab to the correct place
  const elWidth = (ev.currentTarget as HTMLElement).offsetWidth
  if (state.tabs.length == 0) {
    state.tabs.push([data])
  } else if (state.tabs.length == 1) {
    if (offX > elWidth / 3 && offX < (2 * elWidth) / 3) {
      state.tabs[0].push(data)
    } else {
      // Move the middle tab in the other direction
      if (offX < elWidth / 3) {
        state.tabs.unshift([data])
      } else {
        state.tabs.push([data])
      }
    }
  } else {
    if (offX < elWidth / 2) {
      state.tabs[0].push(data)
    } else {
      state.tabs[1].push(data)
    }
  }

  // Remove highlight from dropzones
  state.highlightLeft = false
  state.highlightRight = false
}

const d = ref<HTMLElement | null>(null)
onMounted(() => {
  d.value?.addEventListener("dragover", dragOver)
  d.value?.addEventListener("dragleave", dragLeave)
  d.value?.addEventListener("drop", drop)
})
</script>

<template>
  <!-- MainView contains one or more TabbedViews -->
  <div class="container" ref="d">
    <!-- Display each tab -->
    <TabbedView
      v-for="(tab, n) in state.tabs"
      :key="n"
      :tabs="tab"
      :active="true"
    ></TabbedView>

    <div v-if="state.tabs.length == 0" class="help">
      <!-- Show background if there are no tabs -->
      <p class="help body-large" v-for="i in state.tabs" :key="i[0]">
        Double click or drag a channel to get started
      </p>
    </div>

    <div class="highlight-left" v-show="state.highlightLeft"></div>
    <div class="highlight-right" v-show="state.highlightRight"></div>
  </div>
</template>

<style scoped>
.container {
  height: 100%;
  width: 100%;
  position: relative;

  display: flex;
  flex-direction: row;
}

.help {
  height: 100%;
  width: 100%;
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.highlight-left,
.highlight-right {
  background-color: var(--md-on-primary);
  opacity: 0.1;
  height: 100%;
  width: 50%;
  position: absolute;
}

.highlight-left {
  top: 0;
  left: 0;
}

.highlight-right {
  top: 0;
  left: 50%;
}
</style>
