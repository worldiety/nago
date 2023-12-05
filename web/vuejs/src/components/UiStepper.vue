<script lang="ts" setup>
import {LiveStepper} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveStepper;
  ws: WebSocket;
}>();
</script>

<template>
  <ol class="items-center space-y-4 sm:flex sm:space-x-8 sm:space-y-0 rtl:space-x-reverse">
    <li
        v-for="(step,index) in props.ui.steps.value"
        :class="index==props.ui.selectedIndex.value?'text-blue-600 dark:text-blue-500 dark:border-blue-500':'text-gray-500 dark:text-gray-400'"
        class="flex items-center space-x-2.5 rtl:space-x-reverse">
        <span
            :class="index==props.ui.selectedIndex.value?'border-blue-600':'border-gray-500'"
            class="flex items-center justify-center w-8 h-8 border  rounded-full shrink-0 ">
            {{ step.number.value }}
        </span>
      <span>
            <h3 class="font-medium leading-tight">{{ step.caption.value }}</h3>
            <p class="text-sm">{{ step.details.value }}</p>
        </span>
    </li>
  </ol>
</template>
