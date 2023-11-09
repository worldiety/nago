<script lang="ts" setup>
import type { TableColumnHeader } from '@/shared/model';
import UiGeneric from '@/components/UiGeneric.vue';

const props = defineProps<{
    ui: TableColumnHeader;
}>();
</script>

<template>
    <th scope="col" class="px-6 py-3">
        <ui-generic v-for="ui in props.ui.views" :ui="ui" />
    </th>
</template>
