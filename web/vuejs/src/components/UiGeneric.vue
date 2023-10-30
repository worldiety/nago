<script lang="ts" setup>

import {UiElement} from "@/shared/model";
import UiText from "@/components/UiText.vue";
import UiButton from "@/components/UiButton.vue";
import UiGrid from "@/components/UiGrid.vue";
import UiInputText from "@/components/UiInputText.vue";
import UiInputFile from "@/components/UiInputFile.vue";
import UiCard from "@/components/UiCard.vue";
import UiTable from "@/components/UiTable.vue";

const props = defineProps<{
  ui: UiElement,
}>();

</script>

<template>
  <ui-text v-if="props.ui.type === 'AttributedText'" :ui="props.ui"/>
  <ui-text v-else-if="props.ui.type === 'Text'" :ui="props.ui"/>
  <ui-button v-else-if="props.ui.type === 'Button'" :ui="props.ui"/>
  <ui-grid v-else-if="props.ui.type === 'Grid'" :ui="props.ui"/>
  <ui-table v-else-if="props.ui.type === 'Table'" :ui="props.ui"/>
  <ui-card v-else-if="props.ui.type === 'Card'" :ui="props.ui"/>
  <ui-input-text v-else-if="props.ui.type === 'InputText'" :ui="props.ui"/>
  <ui-input-file v-else-if="props.ui.type === 'InputFile'" :ui="props.ui"/>
  <div v-else>[Unknown view type {{ props.ui.type }}]</div>
</template>