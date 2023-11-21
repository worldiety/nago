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
    <v-switch
        class="mt-2"
        :name="props.ui.id"
        v-model="props.ui.value"
        true-value="true"
        false-value="false"
        clearable
        color="primary"
        persistent-hint
        hide-details="auto"
        :error-messages="props.ui.error"
        :hint="props.ui.hint"
        :label="props.ui.label"
    ></v-switch>
  </v-responsive>

</template>
