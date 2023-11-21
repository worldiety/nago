<script lang="ts" setup>
import type { TextElement } from '@/shared/model';

const props = defineProps<{
    ui: TextElement;
}>();
</script>

<template>
  <v-card-text>{{props.ui.value}}</v-card-text>
</template>
