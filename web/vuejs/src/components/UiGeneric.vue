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

const props = defineProps<{
  ui: UiElement;
}>();
</script>

<template>
  <ui-scaffold v-if="props.ui.type === 'Page'" :ui="props.ui"/>
  <ui-list-view v-else-if="props.ui.type === 'ListView'" :ui="props.ui"/>
  <ui-card-view v-else-if="props.ui.type === 'CardView'" :ui="props.ui"/>
  <ui-card-text v-else-if="props.ui.type === 'CardText'" :ui="props.ui"/>
  <ui-form v-else-if="props.ui.type === 'Form'" :ui="props.ui"/>
  <ui-text-field v-else-if="props.ui.type === 'TextField'" :ui="props.ui"/>
  <ui-file-field v-else-if="props.ui.type === 'FileUploadField'" :ui="props.ui"/>
  <ui-select-field v-else-if="props.ui.type === 'SelectField'" :ui="props.ui"/>
  <ui-text v-else-if="props.ui.type === 'AttributedText'" :ui="props.ui"/>
  <ui-text v-else-if="props.ui.type === 'Text'" :ui="props.ui"/>
  <ui-button v-else-if="props.ui.type === 'Button'" :ui="props.ui"/>
  <ui-grid v-else-if="props.ui.type === 'Grid'" :ui="props.ui"/>
  <ui-table v-else-if="props.ui.type === 'Table'" :ui="props.ui"/>
  <ui-card v-else-if="props.ui.type === 'Card'" :ui="props.ui"/>
  <!-- <ui-navbar v-else-if="props.ui.type === 'Navbar'" :ui="props.ui" /> -->

  <div v-else>[Unknown view type '{{ props.ui.type }}']</div>
</template>
