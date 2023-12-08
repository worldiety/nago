<script lang="ts" setup>
import type {UiDescription} from '@/shared/model';
import type {Ref} from 'vue';
import {inject} from 'vue';
import {invokeFunc, LiveImage, LivePage} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveImage;
  ws: WebSocket
  page: LivePage
}>();

const ui: Ref<UiDescription> = inject('ui')!;

function onClick() {
  invokeFunc(props.ws, props.ui.action)
}

function getSource():string{
  if (props.ui.url.value==="/api/v1/download"){
    return props.ui.url.value+"?page="+props.page.token+"&download="+props.ui.downloadToken.value
  }

  return props.ui.url.value
}

</script>

<template>


  <figure class="max-w-lg mx-auto">
    <img class="h-auto max-w-full rounded-lg" :src="getSource()" :alt="props.ui.caption.value">
    <figcaption v-if="props.ui.caption.value" class="mt-2 text-sm text-center text-gray-500 dark:text-gray-400">
      {{ props.ui.caption.value }}
    </figcaption>
  </figure>

</template>
