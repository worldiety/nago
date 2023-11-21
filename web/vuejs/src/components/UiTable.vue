<script lang="ts" setup>
import type {TableElement, TableListResponse} from '@/shared/model';
import {ref} from 'vue';

import {VDataTable} from 'vuetify/labs/VDataTable'
import router from "@/router";

const props = defineProps<{
  ui: TableElement;
}>();

const tableModel = ref<TableListResponse>({"headers": [], "rows": []});

async function init(): Promise<void> {
  if (props.ui.links.list != null) {
    tableModel.value = await fetch(props.ui.links.list).then((r) => r.json());
  }
  console.log("meh", tableModel.value.headers.values())
}

init();


function tableHeaders(): any {
  return tableModel.headers
}

function tableRows(): any {
  return tableModel.rows
}

function handleClick(item, row) {
  let action = row.item["_action"]
  if (action != null && action.type ==="Redirect"){
    router.push(action.target)
  }
  console.log()
}

</script>

<template>


  <v-data-table
      :headers="tableModel.headers"
      :items="tableModel.rows"
      height="400"
      class="elevation-1"
      @click:row="handleClick"

      items-per-page-text="Zeilen pro Seite"
      :pageText="'{0}-{1} von {2}'"
  ></v-data-table>
</template>
