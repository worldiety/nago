<script lang="ts" setup>

import { GridElement } from "@/shared/model";
import { computed } from "vue";
import UiGridCell from "@/components/UiGridCell.vue";

const props = defineProps<{
    ui: GridElement,
}>();

const style = computed<string>(() => `
    display: grid;
    grid-template-columns: repeat(${props.ui.columns}, 1fr);
`);

</script>

<template>
    <div :style="style">
        <ui-grid-cell v-for="cell in props.ui.cells" :ui="cell" />
    </div>
</template>