<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {VBox} from "@/shared/livemsg";

const props = defineProps<{
  ui: VBox;
  ws: WebSocket;
}>();
</script>

<template>
  <tr class="px-6 py-4">
    <ui-generic v-for="ui in props.ui.children.value" :ui="ui" :ws="ws"/>

  </tr>
</template>
