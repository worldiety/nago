<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import {LiveGridCell} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveGridCell;
  ws: WebSocket;
}>();

const styleGridColStart = computed<string>(() => {
  if (props.ui.colStart.value > 0) {
    return `col-start-${props.ui.colStart.value} `;
  }

  return '';
});

const styleGridColEnd = computed<string>(() => {
  if (props.ui.colEnd.value > 0) {
    return `col-end-${props.ui.colEnd.value} `;
  }
  return '';
});


const styleGridRowStart = computed<string>(() => {
  if (props.ui.rowStart.value > 0) {
    return `row-start-${props.ui.rowStart.value} `;
  }

  return '';
});

const styleGridRowEnd = computed<string>(() => {
  if (props.ui.rowEnd.value > 0) {
    return `row-end-${props.ui.rowEnd.value} `;
  }
  return '';
});

const styleGridColSpan = computed<string>(() => {
  if (props.ui.rowEnd.value > 0) {
    return `row-end-${props.ui.rowEnd.value} `;
  }
  return '';
});

const style = computed(
    () => `
    ${styleGridColStart.value}
    ${styleGridColEnd.value}
    ${styleGridRowStart.value}
    ${styleGridRowEnd.value}
`
);
</script>

<template>
  <div :class="style">
    <ui-generic :ui="props.ui.body.value" :ws="ws"/>
  </div>
</template>
