<script lang="ts" setup>
import uiComponentsMap from '@/shared/uiComponents';
import type { UiElement } from '@/shared/model/uiElement';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: UiElement;
	page: LivePage;
}>();
</script>

<template>
	<component
		:is="uiComponentsMap.get(props.ui.type)"
		v-if="uiComponentsMap.has(props.ui.type)"
		:ui="props.ui"
		:page="props.page"
	/>
	<div v-else>[Unknown view type '{{ props.ui.type }}']</div>
</template>
