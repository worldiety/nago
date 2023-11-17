<script lang="ts" setup>
import type {ListView, ListViewList, LVLinks} from '@/shared/model';
import {ref} from 'vue';
import * as url from "url";

const props = defineProps<{
  ui: ListView;
}>();

const listModel = ref<ListViewList>({});
const menuRequired = ref<boolean>(false);
const canDelete = ref<boolean>(false);
const showDeleteDialog = ref<boolean>(false);
const canSelect = ref<boolean>(false);
const selectedItems = ref<Set<string>>(new Set<string>());

async function init(): Promise<void> {
  if (props.ui.links.list!=null){
    listModel.value = await fetch(props.ui.links.list).then((r) => r.json());
  }
  console.log(listModel.value)
  if (props.ui.links.delete != null) {
    menuRequired.value = true;
    canDelete.value = true;
    canSelect.value = true;
  }
}

async function deleteItems():Promise<void> {
  await fetch(props.ui.links.delete,{
    method:"POST",
    body:JSON.stringify({"identifiers":Array.from(selectedItems.value.values())})
  })
  await init()
}

function updateSelection(id: string,event: any){
  const checked = event.target.checked;
  console.log(id+" => "+checked)
  if (checked){
    selectedItems.value.add(id);
  }else{
    selectedItems.value.delete(id);
  }

  console.log(Array.from(selectedItems.value.values()))
}

init();
</script>

<template>

  <v-dialog
      v-model="showDeleteDialog"
      persistent
      width="auto"
  >
    <v-card>
      <v-card-text>
        Die ausgewählten {{selectedItems.size}} Einträge wirklich löschen?
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary"  @click="showDeleteDialog = false">Abbrechen</v-btn>
        <v-btn color="primary"  @click="showDeleteDialog = false;deleteItems()">Einträge löschen</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>


  <v-list lines="one">


    <div v-if="menuRequired" class="d-flex flex-row-reverse mb-6">
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn variant="plain" icon="mdi-dots-vertical" v-bind="props"></v-btn>
        </template>

        <v-list>
          <v-list-item v-if="canDelete" @click="showDeleteDialog=true">

            <v-list-item-title>Löschen</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
      <v-divider></v-divider>
    </div>


    <v-list-item
        v-for="(item,idx) in listModel.data"
        :key="idx"

        color="primary"
        variant="plain"
    >
      <template v-if="canSelect" v-slot:prepend="{ isActive }">
        <v-list-item-action start>
          <v-checkbox-btn :model-value="isActive" @change="updateSelection(item.id,$event)"></v-checkbox-btn>



        </v-list-item-action>
      </template>

      <v-list-item-title>{{item.title}}</v-list-item-title>


    </v-list-item>
  </v-list>

</template>
