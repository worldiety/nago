<script lang="ts" setup>
import type {UiDescription} from '@/shared/model';
import type {Ref} from 'vue';
import {computed, inject} from 'vue';
import {CallServerFunc, invokeFunc, LiveButton} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveButton;
  ws: WebSocket;
}>();

const ui: Ref<UiDescription> = inject('ui')!;

function onClick() {
  invokeFunc(props.ws,props.ui.action)
}

const clazz = computed<string>(() => {
  switch (props.ui.color.value) {
    case 'primary':
      return 'btn-primary';
    case 'secondary':
      return 'btn-secondary';
    case 'subtile':
      return 'btn-subtile';
    case 'destructive':
      return 'btn-destructive';
    default:
      return 'btn-default';
  }
})

const iconOnly = computed<boolean>(() => {
  return props.ui.caption.value == "" && props.ui.preIcon.value != ""
})

</script>

<style>
.btn-primary, .btn-secondary, .btn-subtile, .btn-destructive {
  @apply inline-flex items-center justify-center  px-5 py-2.5;
  @apply rounded-lg ;
  @apply text-center text-sm font-medium;
  @apply focus:ring-4 focus:outline-none;
}

.btn-primary {
  @apply focus:ring-blue-300;
  @apply bg-blue-700 hover:bg-blue-800;
  @apply text-white;
  @apply dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800;
}

.btn-secondary {
  @apply bg-gray-800 hover:bg-gray-800;
  @apply focus:ring-gray-300;
  @apply text-white;
  @apply dark:bg-gray-800 dark:hover:bg-gray-800 dark:focus:ring-gray-700 dark:border-gray-700;
}

.btn-subtile {
  @apply hover:bg-gray-100;
  @apply focus:ring-gray-300;
  @apply text-gray-900 border;
  @apply dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:hover:border-gray-600 dark:focus:ring-gray-700;
}

.btn-destructive {
  @apply focus:ring-gray-300;
  @apply focus:outline-none text-white bg-red-700 hover:bg-red-800 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-900;
}

.btn-default {
  @apply py-2.5 px-5 me-2 mb-2 text-sm font-medium text-gray-900 focus:outline-none rounded-lg hover:bg-gray-100 hover:text-blue-700 focus:z-10 focus:ring-4 focus:ring-gray-200 dark:focus:ring-gray-700  dark:text-gray-400 dark:border-gray-600 dark:hover:text-white dark:hover:bg-gray-700
}
</style>

<template>
  <button
      :class="clazz"
      @click="onClick" :disabled="props.ui.disabled.value">
    <svg v-inline class="w-3.5 h-3.5 me-2" v-html="props.ui.preIcon.value" v-if="props.ui.preIcon.value && !iconOnly"></svg>
    {{ props.ui.caption.value }}
    <span class="w-3.5 h-3.5 ms-2" v-html="props.ui.postIcon.value" v-if="props.ui.postIcon.value"></span>
    <svg v-inline class="w-4 h-4" v-if="iconOnly" v-html="props.ui.preIcon.value"></svg>
  </button>


</template>
