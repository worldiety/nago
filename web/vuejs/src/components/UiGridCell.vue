<script lang="ts" setup>

import { GridCellElement} from "@/shared/model";
import UiGeneric from "@/components/UiGeneric.vue";
import { computed } from "vue";

const props = defineProps<{
    ui: GridCellElement,
}>();



const styleGridColSpan = computed<string>(() => {
  if (props.ui.colSpan > 0) {
    return `col-span-${props.ui.colSpan} `;
  }
  return "";
});

const styleGridRowSpan = computed<string>(() => {
  if (props.ui.rowSpan > 0) {
    return `row-span-${props.ui.rowSpan} `;
  }
  return "";
});

const style = computed(() => `
    ${styleGridRowSpan.value}
    ${styleGridColSpan.value}
`);

</script>

<template>
    <div :class="style">
        <ui-generic v-for="ui in props.ui.views" :ui="ui" />
    </div>
</template>