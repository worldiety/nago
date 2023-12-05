<script lang="ts" setup>
import type {UiElement} from '@/shared/model';
import UiText from '@/components/UiText.vue';
import UiButton from '@/components/UiButton.vue';
import UiGrid from '@/components/UiGrid.vue';
import UiInputText from '@/components/UiInputText.vue';
import UiInputFile from '@/components/UiInputFile.vue';
import UiCard from '@/components/UiCard.vue';
import UiTable from '@/components/UiTable.vue';
import UiScaffold from '@/components/UiScaffold.vue';
import UiListView from '@/components/UiListView.vue';
import UiForm from "@/components/UiForm.vue";
import UiTextField from "@/components/UiTextField.vue";
import UiFileField from "@/components/UiFileField.vue";
import UiSelectField from "@/components/UiSelectField.vue";
import UiCardView from "@/components/UiCardView.vue";
import UiCardText from "@/components/UiCardText.vue";
import UiSwitchField from "@/components/UiSwitchField.vue";
import UiDateField from "@/components/UiDateField.vue";
import UiCardMetricText from "@/components/UiCardMetricText.vue";
import UiSVG from "@/components/UiSVG.vue";
import UiVBox from "@/components/UiVBox.vue";
import UiHBox from "@/components/UiHBox.vue";
import UiDivider from "@/components/UiDivider.vue";
import UiDialog from "@/components/UiDialog.vue";
import UiToggle from "@/components/UiToggle.vue";
import UiStepper from "@/components/UiStepper.vue";
import UiTextArea from "@/components/UiTextArea.vue";
import UiChip from "@/components/UiChip.vue";

const props = defineProps<{
  ui: UiElement;
  ws: WebSocket;
}>();
</script>

<template>
  <ui-scaffold v-if="props.ui.type === 'Scaffold'" :ui="props.ui" :ws="props.ws"/>
  <ui-v-box v-else-if="props.ui.type === 'VBox'" :ui="props.ui" :ws="props.ws"/>
  <ui-h-box v-else-if="props.ui.type === 'HBox'" :ui="props.ui" :ws="props.ws"/>
  <ui-chip v-else-if="props.ui.type === 'Chip'" :ui="props.ui" :ws="props.ws"/>
  <ui-dialog v-else-if="props.ui.type === 'Dialog'" :ui="props.ui" :ws="props.ws"/>
  <ui-list-view v-else-if="props.ui.type === 'ListView'" :ui="props.ui" :ws="props.ws"/>
  <ui-divider v-else-if="props.ui.type === 'Divider'" :ui="props.ui" :ws="props.ws"/>
  <ui-s-v-g v-else-if="props.ui.type === 'SVG'" :ui="props.ui"/>
  <ui-stepper v-else-if="props.ui.type === 'Stepper'" :ui="props.ui" :ws="props.ws"/>

  <ui-form v-else-if="props.ui.type === 'Form'" :ui="props.ui"/>
  <ui-text-field v-else-if="props.ui.type === 'TextField'" :ui="props.ui" :ws="props.ws"/>
  <ui-text-area v-else-if="props.ui.type === 'TextArea'" :ui="props.ui" :ws="props.ws"/>
  <ui-toggle v-else-if="props.ui.type === 'Toggle'" :ui="props.ui" :ws="props.ws"/>
  <ui-date-field v-else-if="props.ui.type === 'DateField'" :ui="props.ui"/>
  <ui-file-field v-else-if="props.ui.type === 'FileUploadField'" :ui="props.ui"/>
  <ui-select-field v-else-if="props.ui.type === 'SelectField'" :ui="props.ui"/>
  <ui-text v-else-if="props.ui.type === 'Text'" :ui="props.ui" :ws="props.ws"/>
  <ui-button v-else-if="props.ui.type === 'Button'" :ui="props.ui" :ws="props.ws"/>
  <ui-grid v-else-if="props.ui.type === 'Grid'" :ui="props.ui" :ws="props.ws"/>
  <ui-table v-else-if="props.ui.type === 'Table'" :ui="props.ui" :ws="props.ws"/>
  <ui-card v-else-if="props.ui.type === 'Card'" :ui="props.ui" :ws="props.ws"/>
  <!-- <ui-navbar v-else-if="props.ui.type === 'Navbar'" :ui="props.ui" /> -->

  <div v-else>[Unknown view type '{{ props.ui.type }}']</div>
</template>
