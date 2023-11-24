<script lang="ts" setup>
import type {SVGElement} from '@/shared/model';

const props = defineProps<{
  ui: SVGElement;
}>();
</script>

<template>
  <v-responsive
      class="mx-auto"
      :max-width="props.ui.maxWidth"
  >
    <div v-html="props.ui.svg"></div>
  </v-responsive>
</template>
