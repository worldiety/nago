<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import type { LiveGridCell, LivePage } from '@/shared/model';

const props = defineProps<{
  ui: LiveGridCell;
  page: LivePage;
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
  let tmp = ""
  if (props.ui.colSpan.value > 0) {
    tmp+= `col-span-${props.ui.colSpan.value} `;
  }

  if (props.ui.smColSpan.value > 0) {
    tmp+= ` sm:col-span-${props.ui.smColSpan.value} `;
  }

  if (props.ui.mdColSpan.value > 0) {
    tmp+= ` md:col-span-${props.ui.mdColSpan.value} `;
  }

  if (props.ui.lgColSpan.value > 0) {
    tmp+= ` lg:col-span-${props.ui.lgColSpan.value} `;
  }

  return tmp;
});

const style = computed(
    () => `
    ${styleGridColStart.value}
    ${styleGridColEnd.value}
    ${styleGridRowStart.value}
    ${styleGridRowEnd.value}
    ${styleGridColSpan.value}
`
);
</script>

<template>
  <div :class="style">
    <ui-generic :ui="props.ui.body.value" :page="page"/>
  </div>
</template>
