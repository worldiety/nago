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
    <v-file-input
        chips
        :multiple="props.ui.fileMultiple"
        class="mt-2"
        :name="props.ui.id"
        :accept="props.ui.fileAccept"
        clearable
        persistent-hint
        hide-details="auto"
        :error-messages="props.ui.error"
        :hint="props.ui.hint"
        :label="props.ui.label"
        :disabled="props.ui.disabled"
    ></v-file-input>
  </v-responsive>

</template>
