<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { Page } from "@/shared/protocol/ora/page";

const props = defineProps<{
	ui: Page;
}>();

</script>

<template>
	<!-- Modals -->
	<div class="fixed inset-0 pointer-events-none z-40">
		<UiGeneric v-for="(modal, index) in ui.modals.v" :key="index" :ui="modal"/>
	</div>

	<!-- Page content -->
	<div class="fixed inset-0 bg-white dark:bg-darkmode-gray">
		<UiGeneric v-if="props.ui.body.v" :ui="props.ui.body.v"  />
	</div>
</template>
