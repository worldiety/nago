<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {HBox, VBox} from "@/shared/livemsg";

const props = defineProps<{
  ui: HBox;
  ws: WebSocket;
}>();
</script>

<template>
   <div class="grid grid-flow-col grid-cols-auto gap-x-3">

    <ui-generic v-for="ui in props.ui.children.value" :ui="ui" :ws="ws"/>
  </div>
</template>
