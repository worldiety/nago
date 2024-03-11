<script lang="ts" setup>
import type { Divider, LivePage } from '@/shared/model';

const props = defineProps<{
  ui: Divider;
  page: LivePage;
}>();
</script>

<template>
  <hr
      class="border-t-1 opacity-100 dark:opacity-50" />
</template>
