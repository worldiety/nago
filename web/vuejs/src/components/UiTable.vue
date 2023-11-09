<script lang="ts" setup>
import type { TableElement } from '@/shared/model';
import { computed } from 'vue';
import UiTableColumnHeader from '@/components/UiTableColumnHeader.vue';
import UiTableRow from '@/components/UiTableRow.vue';

const props = defineProps<{
    ui: TableElement;
}>();

//TODO we get into trouble using tailwind pre-processor here
const style = computed<string>(
    () => `

`
);
</script>

<template>
    <table class="w-full text-left text-sm text-gray-500">
        <thead class="bg-gray-50 text-xs uppercase text-gray-700">
            <tr>
                <ui-table-column-header v-for="cell in props.ui.columnHeaders" :ui="cell" />
            </tr>
        </thead>

        <tbody>
            <ui-table-row v-for="cell in props.ui.rows" :ui="cell" />
        </tbody>
    </table>
</template>
