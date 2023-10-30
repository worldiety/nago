<script lang="ts" setup>

import { GridElement } from "@/shared/model";
import { computed } from "vue";
import UiGridCell from "@/components/UiGridCell.vue";

const props = defineProps<{
    ui: GridElement,
}>();

//TODO we get into trouble using tailwind pre-processor here
const style = computed<string>(() => `
    grid
    grid-cols-${props.ui.columns}
    grid-rows-${props.ui.rows}
    gap-[${props.ui.gap}]
`);

</script>

<template>
    <div :class="style">
        <ui-grid-cell v-for="cell in props.ui.cells" :ui="cell" />
    </div>
</template>