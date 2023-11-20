<script lang="ts" setup>
import type {FormField, InputTextElement} from '@/shared/model';

const props = defineProps<{
  ui: FormField;
}>();


</script>

<template>

  <v-responsive
      class="mx-auto"
      max-width="344"
  >
    <v-select
        chips
        class="mt-2"
        :name="props.ui.id"
        :items="props.ui.selectItems"
        v-model="props.ui.selectValues"
        item-title="title"
        item-value="value"
        :multiple="props.ui.selectMultiple"
        :disabled="props.ui.disabled"
        clearable
        persistent-hint
        hide-details="auto"
        :error-messages="props.ui.error"
        :hint="props.ui.hint"
        :label="props.ui.label"
    ></v-select>
  </v-responsive>

</template>
