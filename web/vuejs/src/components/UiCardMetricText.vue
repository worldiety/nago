<script lang="ts" setup>
import type {CardMetricText, TextElement} from '@/shared/model';

const props = defineProps<{
    ui: CardMetricText;
}>();
</script>

<template>

  <v-card-text class="py-0">
    <v-row align="center" no-gutters>
      <v-col
          class="text-h3"
          cols="8"
      >
        {{props.ui.value}}
      </v-col>
      <v-col v-if="props.ui.icon" cols="4" class="text-right">
        <v-icon
            :color="props.ui.icon.color"
            :icon="props.ui.icon.name"
            size="66"
        ></v-icon>
      </v-col>
    </v-row>
  </v-card-text>

</template>
