<script lang="ts" setup>

import { GridCellElement} from "@/shared/model";
import UiGeneric from "@/components/UiGeneric.vue";
import { computed } from "vue";

const props = defineProps<{
    ui: GridCellElement,
}>();

const styleGridStart = computed<string>(() => {
    if (props.ui.start > 0) {
        return `grid-column-start: ${props.ui.start};`;
    }
    return "";
});
const styleGridEnd = computed<string>(() => {
    if (props.ui.end > 0) {
        return `grid-column-end: ${props.ui.end};`;
    }
    return "";
});
const style = computed(() => `
    ${styleGridStart.value}
    ${styleGridEnd.value}
`);

</script>

<template>
    <div :style="style">
        <ui-generic v-for="ui in props.ui.views" :ui="ui" />
    </div>
</template>