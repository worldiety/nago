<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {Divider, VBox} from "@/shared/livemsg";

const props = defineProps<{
  ui: Divider;
  ws: WebSocket;
}>();
</script>

<template>
  <hr
      class="border-t-1 opacity-100 dark:opacity-50" />
</template>
