<script setup lang="ts">
import { onMounted, ref } from "vue"

const props = defineProps<{ text: string; icon: string }>()

function dragStart(ev: Event) {
  ;(ev as DragEvent).dataTransfer?.setData(
    "text/plain",
    (ev.target as HTMLElement).innerHTML
  )
}

const d = ref<HTMLElement | null>(null)
onMounted(() => {
  d.value?.addEventListener("dragstart", dragStart)
})
</script>

<template>
  <div class="title-medium" ref="d" draggable="true">
    <span class="icon material-symbols-rounded">{{ props.icon }}</span>
    {{ props.text }}
  </div>
</template>

<style scoped>
div {
  padding: 6px;
  margin-top: 5px;
  border-radius: 3px;
  transition: all 0.1s;

  display: flex;
  flex-direction: row;
}

div:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.icon {
  transform: scale(0.8);
  margin-right: 6px;
}
</style>
