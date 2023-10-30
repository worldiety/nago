<script lang="ts" setup>

import {TableColumnHeader, TableRow} from "@/shared/model";
import UiGeneric from "@/components/UiGeneric.vue";
import UiTableCell from "@/components/UiTableCell.vue";

const props = defineProps<{
  ui: TableRow,
}>();


</script>

<template>
  <tr  class="bg-white border-b">
    <ui-table-cell v-for="ui in props.ui.columns" :ui="ui"/>
  </tr>

</template>