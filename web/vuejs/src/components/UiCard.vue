<script lang="ts" setup>
import type {UiDescription} from '@/shared/model';
import type {Ref} from 'vue';
import {inject} from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import {invokeFunc, LiveCard, LivePage} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveCard;
  ws: WebSocket
  page: LivePage
}>();

const ui: Ref<UiDescription> = inject('ui')!;

function onClick() {
  invokeFunc(props.ws, props.ui.action)
}

</script>

<template>


  <div @click="onClick" :class="props.ui.action.value>0?'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700':''"
       class="block max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow  dark:bg-gray-800 dark:border-gray-700 ">
    <ui-generic v-for="ui in props.ui.children.value" :ui="ui" :ws="ws" :page="page"/>
  </div>

</template>
