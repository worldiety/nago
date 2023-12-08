<script lang="ts" setup>
import {invokeFunc, LivePage, LiveToggle, SetServerProperty} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveToggle;
  ws: WebSocket;
  page: LivePage
}>();

function onClick() {
  const setSrvProp: SetServerProperty = {
    type: "setProp",
    id: props.ui.checked.id,
    value: props.ui.checked.value
  }
  props.ws.send(JSON.stringify(setSrvProp))

  invokeFunc(props.ws,props.ui.onCheckedChanged)
}

</script>

<template>

  <label class="relative inline-flex items-center cursor-pointer">
    <input @change="onClick" v-model="props.ui.checked.value" type="checkbox" value="" class="sr-only peer" :checked="props.ui.checked.value" :disabled="props.ui.disabled.value">
    <span v-if="ui.disabled.value"
        class="w-11 h-6 bg-gray-200 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-gray-400"></span>
    <span v-else class="w-11 h-6 bg-gray-200 rounded-full peer peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 dark:bg-gray-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></span>
    <span class="ms-3 text-sm font-medium text-gray-900 dark:text-gray-300">{{props.ui.label.value}}</span>
  </label>

</template>
