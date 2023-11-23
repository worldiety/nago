<script lang="ts" setup>
import type {TableElement, TableListResponse} from '@/shared/model';
import {ref} from 'vue';

import router from "@/router";
import {useHttp} from "@/shared/http";

const props = defineProps<{
  ui: TableElement;
}>();

const tableModel = ref<TableListResponse>({"headers": [], "rows": []});

async function init(): Promise<void> {
  const http = useHttp();
  if (props.ui.links.list != null) {
    tableModel.value = await http.request(props.ui.links.list).then((r) => r.json());
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
  if (action != null && action.type === "Redirect") {
    router.push(action.target)
  }
  console.log()
}

function hasHover(): boolean {
  if (tableModel.value != null && tableModel.value.rows.length > 0) {
    let action = tableModel.value.rows[0]["_action"]
    if (action != null && action.type === "Redirect") {
      return true
    }
  }

  return false
}

const search = ref<string>("")
</script>

<template>

  <v-card
      flat

  >

    <template v-slot:text>
      <v-row>
        <v-text-field
            v-model="search"
            label="Stichwort Tabellenfilter"
            prepend-inner-icon="mdi-magnify"
            single-line
            clearable
            variant="outlined"
            hide-details
        ></v-text-field>

        <v-btn
            color="primary"
            dark
            class="mb-2"
            v-bind="props"
        >
          New Item
        </v-btn>
      </v-row>
    </template>


    <v-data-table
        :headers="tableModel.headers"
        :items="tableModel.rows"
        height="400"
        v-bind:hover="hasHover()"
        @click:row="handleClick"
        :search="search"

        items-per-page-text="Zeilen pro Seite"
        :pageText="'{0}-{1} von {2}'"
    >

      <template v-slot:item.actions="{ item }">
        <v-icon
            size="small"
            class="me-2"
            @click="editItem(item)"
        >
          mdi-pencil
        </v-icon>
        <v-icon
            size="small"
            @click="deleteItem(item)"
        >
          mdi-delete
        </v-icon>
      </template>
      <template v-slot:no-data>
        <v-btn
            color="primary"
            @click="initialize"
        >
          Reset
        </v-btn>
      </template>

    </v-data-table>
  </v-card>
</template>
