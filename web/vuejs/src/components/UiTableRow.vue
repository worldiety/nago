<script lang="ts" setup>
import type { TableRow } from '@/shared/model';
import { TableColumnHeader } from '@/shared/model';
import UiGeneric from '@/components/UiGeneric.vue';
import UiTableCell from '@/components/UiTableCell.vue';

const props = defineProps<{
    ui: TableRow;
}>();
</script>

<template>
    <tr class="border-b bg-white">
        <ui-table-cell v-for="ui in props.ui.columns" :ui="ui" />
    </tr>
</template>
