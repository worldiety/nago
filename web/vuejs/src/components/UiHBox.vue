<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {HBox, VBox} from "@/shared/livemsg";
import {computed} from "vue";

const props = defineProps<{
  ui: HBox;
  ws: WebSocket;
}>();

const clazz = computed<string>(() => {

  switch (props.ui.alignment.value) {
    case 'flex-center':
      return "flex justify-center"
    case 'flex-left':
      return "flex"
    case 'flex-right':
      return "flex flex-row-reverse"
    case 'grid':
    default:
      return "grid grid-flow-col grid-cols-auto gap-x-3"
  }
})

</script>

<template>
   <div :class="clazz">

    <ui-generic v-for="ui in props.ui.children.value" :ui="ui" :ws="ws"/>
  </div>
</template>
