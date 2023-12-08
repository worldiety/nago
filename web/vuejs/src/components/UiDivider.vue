<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {Divider, LivePage, VBox} from "@/shared/livemsg";

const props = defineProps<{
  ui: Divider;
  ws: WebSocket;
  page: LivePage
}>();
</script>

<template>
  <hr
      class="border-t-1 opacity-100 dark:opacity-50" />
</template>
