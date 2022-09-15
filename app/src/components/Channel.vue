<script setup lang="ts">
import { defineProps, onMounted, ref } from "vue";

const props = defineProps<{ text: string }>();

function dragStart(ev: Event) {
  (ev as DragEvent).dataTransfer?.setData("text/plain", (ev.target as HTMLElement).innerHTML);
}

const d = ref<HTMLElement | null>(null);
onMounted(() => {
  d.value?.addEventListener("dragstart", dragStart)
})
</script>
    
<template>
  <div ref="d" draggable="true">
    {{ props.text }}
  </div>
</template>
    
<style scoped>

</style>
    