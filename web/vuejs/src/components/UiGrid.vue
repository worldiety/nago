<script lang="ts" setup>
import {computed} from 'vue';
import UiGridCell from '@/components/UiGridCell.vue';
import {gapSize2Tailwind, LiveGrid} from "@/shared/livemsg";

const props = defineProps<{
  ui: LiveGrid;
  ws: WebSocket;
}>();

//TODO we get into trouble using tailwind pre-processor here
const style = computed<string>(() => {
      let tmp = "grid"
      if (props.ui.columns.value > 0) {
        tmp += ` grid-cols-${props.ui.columns.value}`
      } else {
        if (props.ui.rows.value>0){
          tmp+=" grid-flow-col"
        }else{
          tmp += " grid-cols-auto"
        }
      }

      if (props.ui.rows.value > 0) {
        tmp += ` grid-rows-${props.ui.rows.value}`
      } else {
        tmp += " grid-rows-auto"
      }


      tmp += " " + gapSize2Tailwind(props.ui.gap.value)

      return tmp

    }
);
</script>

<template>
  <div :class="style" >
    <ui-grid-cell v-for="cell in props.ui.cells.value" :ui="cell" :ws="props.ws"/>
  </div>
</template>
