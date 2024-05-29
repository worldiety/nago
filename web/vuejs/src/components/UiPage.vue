<template>
	<!-- Modals -->
	<div v-for="(modal, index) in props.ui.modals.v" :key="index" class="modal-container fixed inset-0 pointer-events-none" :style="`--modal-z-index: ${index + 40};`">
		<UiGeneric :ui="modal" :is-active-dialog="index === props.ui.modals.v.length - 1" />
	</div>

	<!-- Page content -->
	<div class="content-container bg-white h-full" :class="{'content-container-freezed': anyModalVisible}" :style="`--content-top-offset: ${windowScrollY}px;`">
		<UiGeneric v-if="props.ui.body.v" :ui="props.ui.body.v"  />
	</div>
</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { Page } from "@/shared/protocol/ora/page";
import { nextTick, ref, watch } from 'vue';

const props = defineProps<{
	ui: Page;
}>();

const anyModalVisible = ref<boolean>(false);
const windowScrollY = ref<number>(0);

watch(() => props.ui.modals.v, (newValue) => {
	if (newValue) {
		if (!anyModalVisible.value) {
			windowScrollY.value = window.scrollY * -1;
			anyModalVisible.value = true;
		}
	} else {
		anyModalVisible.value = false;
		nextTick(() => {
			window.scrollTo(0, windowScrollY.value * -1);
		})
	}
});
</script>

<style scoped>
.modal-container {
	z-index: var(--modal-z-index);
}

.content-container.content-container-freezed {
	@apply fixed left-0 right-0;
	top: var(--content-top-offset);
}
</style>
