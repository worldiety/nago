<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { LivePage, VBox } from '@/shared/model';

const props = defineProps<{
  ui: VBox;
  page: LivePage;
}>();
</script>

<template>
   <div class="grid grid-cols-1 gap-y-3">
    <ui-generic v-for="ui in props.ui.children.value" :ui="ui" :page="page"/>
  </div>
</template>
