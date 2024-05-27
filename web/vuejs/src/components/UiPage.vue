<template>
	<!-- Modals -->
	<div v-for="(modal, index) in modalsByTimestamp" :key="index" class="modal-container fixed inset-0 pointer-events-none" :style="`--modal-z-index: ${index + 40};`">
		<UiDialog :ui="modal" />
	</div>

	<!-- Page content -->
	<div class="fixed inset-0 bg-white dark:bg-darkmode-gray">
		<UiGeneric v-if="props.ui.body.v" :ui="props.ui.body.v"  />
	</div>
</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { Page } from "@/shared/protocol/ora/page";
import { computed } from 'vue';
import type { Dialog } from '@/shared/protocol/ora/dialog';
import UiDialog from '@/components/UiDialog.vue';

const props = defineProps<{
	ui: Page;
}>();

const modalsByTimestamp = computed((): Dialog[] => {
	return props.ui.modals.v
		.flatMap((component) => {
			return component.type === 'Dialog' ? [component as Dialog] : [];
		})
		.toSorted((a, b) => {
			if (a.timestamp.v > b.timestamp.v) {
				return 1;
			} else if (a.timestamp.v < b.timestamp.v) {
				return -1;
			}
			return 0;
		});
});
</script>

<style scoped>
.modal-container {
	z-index: var(--modal-z-index);
}
</style>
