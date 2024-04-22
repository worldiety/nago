<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {VBox} from "@/shared/protocol/gen/vBox";

const props = defineProps<{
	ui: VBox;
}>();
</script>

<template>
	<div class="grid grid-cols-1 gap-y-3">
		<ui-generic v-for="ui in props.ui.children.v" :ui="ui"  />
	</div>
</template>
