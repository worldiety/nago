<script lang="ts" setup>
import type { TableCell} from '@/shared/model';
import { TableColumnHeader, TableRow } from '@/shared/model';
import UiGeneric from '@/components/UiGeneric.vue';

const props = defineProps<{
    ui: TableCell;
}>();
</script>

<template>
    <td class="px-6 py-4">
        <ui-generic v-for="ui in props.ui.views" :ui="ui" />
    </td>
</template>
