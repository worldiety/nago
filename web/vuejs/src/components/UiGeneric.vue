<script lang="ts" setup>
import type {UiElement} from '@/shared/model';
import UiText from '@/components/UiText.vue';
import UiButton from '@/components/UiButton.vue';
import UiGrid from '@/components/UiGrid.vue';
import UiCard from '@/components/UiCard.vue';
import UiTable from '@/components/UiTable.vue';
import UiScaffold from '@/components/UiScaffold.vue';
import UiTextField from "@/components/UiTextField.vue";
import UiVBox from "@/components/UiVBox.vue";
import UiHBox from "@/components/UiHBox.vue";
import UiDivider from "@/components/UiDivider.vue";
import UiDialog from "@/components/UiDialog.vue";
import UiToggle from "@/components/UiToggle.vue";
import UiStepper from "@/components/UiStepper.vue";
import UiTextArea from "@/components/UiTextArea.vue";
import UiChip from "@/components/UiChip.vue";
import UiUploadField from "@/components/UiUploadField.vue";
import {LivePage} from "@/shared/livemsg";
import UiImage from "@/components/UiImage.vue";

const props = defineProps<{
  ui: UiElement;
  ws: WebSocket;
  page: LivePage
}>();
</script>

<template>
  <ui-scaffold v-if="props.ui.type === 'Scaffold'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-v-box v-else-if="props.ui.type === 'VBox'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-h-box v-else-if="props.ui.type === 'HBox'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-chip v-else-if="props.ui.type === 'Chip'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-dialog v-else-if="props.ui.type === 'Dialog'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-divider v-else-if="props.ui.type === 'Divider'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-stepper v-else-if="props.ui.type === 'Stepper'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-upload-field v-else-if="props.ui.type === 'FileField'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-image v-else-if="props.ui.type === 'Image'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-text-field v-else-if="props.ui.type === 'TextField'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-text-area v-else-if="props.ui.type === 'TextArea'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-toggle v-else-if="props.ui.type === 'Toggle'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-text v-else-if="props.ui.type === 'Text'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-button v-else-if="props.ui.type === 'Button'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-grid v-else-if="props.ui.type === 'Grid'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-table v-else-if="props.ui.type === 'Table'" :ui="props.ui" :ws="props.ws" :page="props.page"/>
  <ui-card v-else-if="props.ui.type === 'Card'" :ui="props.ui" :ws="props.ws" :page="props.page"/>

  <div v-else>[Unknown view type '{{ props.ui.type }}']</div>
</template>
