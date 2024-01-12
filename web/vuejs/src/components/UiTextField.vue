<script lang="ts" setup>
import {invokeSetProp, invokeTx2, LivePage, LiveTextField, SetServerProperty} from "@/shared/livemsg";
import {invokeFunc} from "@/shared/livemsg";
import {computed} from "vue";

const props = defineProps<{
  ui: LiveTextField;
  ws: WebSocket;
  page: LivePage
}>();

function valueChanged(event: any) {

  props.ui.value.value = event.target.value


  invokeTx2(props.ws,props.ui.value,props.ui.onTextChanged)
}

function isErr(): boolean {
  return props.ui.error.value != ''
}

const labelClass = computed<string>(() => {
  if (props.ui.disabled.value && isErr()) {
    return 'text-red-900 dark:text-red-700'
  }

  if (isErr()) {
    return 'text-red-700 dark:text-red-500'
  }


  return 'text-gray-900 dark:text-white'
})

const inputClass = computed<string>(() => {
  if (props.ui.disabled.value) {
    return 'bg-gray-100 border border-gray-200 text-gray-600 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 cursor-not-allowed dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-gray-400 dark:focus:ring-blue-500 dark:focus:border-blue-500'
  }

  if (isErr()) {
    return 'bg-red-50 border border-red-500 text-red-900 placeholder-red-700 text-sm rounded-lg focus:ring-red-500 dark:bg-gray-700 focus:border-red-500 block w-full p-2.5 dark:text-red-500 dark:placeholder-red-500 dark:border-red-500'
  }



  return 'bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500'

})

</script>

<template>


  <div>
    <label :for="props.ui.id.toString()"
           :class="labelClass"
           class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</label>
    <input :disabled="props.ui.disabled.value" @input="valueChanged" :value="props.ui.value.value" type="text"
           :id="props.ui.id.toString()"

           :class="inputClass"
    >
    <p v-if="isErr()" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
    <p v-if="!isErr()" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
  </div>

</template>
