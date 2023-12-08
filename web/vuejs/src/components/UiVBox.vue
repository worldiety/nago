<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {LivePage, VBox} from "@/shared/livemsg";

const props = defineProps<{
  ui: VBox;
  ws: WebSocket;
  page: LivePage
}>();
</script>

<template>
   <div class="grid grid-cols-1 gap-y-3">

    <ui-generic v-for="ui in props.ui.children.value" :ui="ui" :ws="ws" :page="page"/>
  </div>
</template>
