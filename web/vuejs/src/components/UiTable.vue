<script lang="ts" setup>
import UiGeneric from "@/components/UiGeneric.vue";
import type { LiveTable } from '@/shared/model/liveTable';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
  ui: LiveTable;
  page: LivePage;
}>();

</script>

<template>

  <div class="relative overflow-x-auto shadow-md">
    <table class="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
      <thead v-if="props.ui.headers.value"
             class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
      <tr>
        <th v-for="head in props.ui.headers.value" scope="col" class="px-6 py-3">
          <ui-generic :ui="head.body.value" :page="page"/>
        </th>
      </tr>
      </thead>

      <tbody>
      <tr v-for="row in props.ui.rows.value" class="odd:bg-white odd:dark:bg-gray-900 even:bg-gray-50 even:dark:bg-gray-800 border-b dark:border-gray-700">
        <td v-for="cell in row.cells.value" class="px-6 py-4">
          <ui-generic :ui="cell.body.value" :page="page"/>
        </td>
      </tr>
      </tbody>
    </table>
  </div>
</template>
