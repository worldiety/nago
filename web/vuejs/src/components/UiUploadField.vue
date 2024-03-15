<script lang="ts" setup>
import {computed} from "vue";
import type { LiveUploadField } from '@/shared/model/liveUploadField';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
  ui: LiveUploadField;
  page: LivePage;
}>();

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

function fileInputChanged(e: Event) {
  const item = e.target
  const formData = new FormData()
  for (const file of item.files) {
    formData.append("files", file);
  }

  fetch("/api/v1/upload", {
    method: 'POST', body: formData, headers: {
      "x-page-token": props.page.token,
      "x-upload-token":props.ui.uploadToken.value,
    }
  })
}

</script>

<template>


  <div class="flex items-center justify-center w-full">
    <label :for="props.ui.id.toString()"
           class="flex flex-col items-center justify-center w-full h-64 border-2 border-gray-300 border-dashed rounded-lg cursor-pointer bg-gray-50 dark:hover:bg-bray-800 dark:bg-gray-700 hover:bg-gray-100 dark:border-gray-600 dark:hover:border-gray-500 dark:hover:bg-gray-600">
      <div class="flex flex-col items-center justify-center pt-5 pb-6">
        <svg class="w-8 h-8 mb-4 text-gray-500 dark:text-gray-400" aria-hidden="true" xmlns="http://www.w3.org/2000/svg"
             fill="none" viewBox="0 0 20 16">
          <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"/>
        </svg>
        <p class="mb-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ props.ui.label.value }}</p>
      </div>
      <input @change="fileInputChanged" :disabled="props.ui.disabled.value" :id="props.ui.id.toString()" type="file"
             class="hidden" :multiple="props.ui.multiple.value" :accept="props.ui.filter.value"/>
      <p v-if="isErr()" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>

    </label>
  </div>

</template>
