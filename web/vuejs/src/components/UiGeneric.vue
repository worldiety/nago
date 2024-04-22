<script lang="ts" setup>
import uiComponentsMap from '@/shared/uiComponents';
import {Component as OraComponent} from "@/shared/protocol/gen/component";

const props = defineProps<{
	ui: OraComponent; // if we don't rename the import, Vue confuses everything up
}>();
</script>

<template>
	<component
		:is="uiComponentsMap.get(props.ui.type)"
		v-if="props.ui && uiComponentsMap.has(props.ui.type)"
		:ui="props.ui"
	/>
	<div v-else-if="!props.ui">ui is null</div>
	<div v-else>[Unknown view type '{{ props.ui.type }}']</div>
</template>
