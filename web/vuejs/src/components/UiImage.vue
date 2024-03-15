<script lang="ts" setup>
import type {Ref} from 'vue';
import {inject} from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveImage } from '@/shared/model/liveImage';
import type { LivePage } from '@/shared/model/livePage';
import type { UiDescription } from '@/shared/model/uiDescription';

const props = defineProps<{
  ui: LiveImage;
  page: LivePage;
}>();

const networkStore = useNetworkStore();
const ui: Ref<UiDescription> = inject('ui')!;

function onClick() {
  networkStore.invokeFunc(props.ui.action);
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
