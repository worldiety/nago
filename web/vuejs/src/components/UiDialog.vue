<template>
	<div v-if="ui.visible.v" class="relative flex justify-center items-center pointer-events-auto bg-black bg-opacity-60 h-full">
		<div class="text-black dark:text-white rounded-lg shadow-md" :class="dialogClass" @click.stop>
			<!-- Dialog header -->
			<div class="flex justify-start items-center gap-x-2 bg-[#F9F9F9] dark:bg-black rounded-t-lg px-6 py-3">
				<div v-html="ui.icon.v" class="w-6 *:h-full"></div>
				<p class="font-bold">{{ ui.title.v }}</p>
			</div>

			<!-- Dialog body -->
			<div class="bg-white dark:bg-[#2B2B2B] pt-3.5 px-6 pb-6" :class="{'rounded-b-lg': !ui.footer.v}">
				<UiGeneric :ui="ui.body.v" />
			</div>

			<div v-if="ui.footer.v" class="bg-white dark:bg-[#2B2B2B] rounded-b-lg px-6 pb-6">
				<hr class="border-[#E2E2E2] dark:border-[#848484] pb-6" />
				<!-- Dialog footer -->
				<UiGeneric :ui="ui.footer.v" />
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import type { Dialog } from '@/shared/protocol/ora/dialog';
import UiGeneric from '@/components/UiGeneric.vue';
import { computed } from 'vue';
import { ElementSize } from '@/shared/protocol/ora/elementSize';

const props = defineProps<{
	ui: Dialog;
}>();

const dialogClass = computed((): string => {
	switch (props.ui.size.v) {
		case ElementSize.SIZE_AUTO:
			return 'w-auto';
		case ElementSize.SIZE_SMALL:
			return 'w-[25rem]';
		case ElementSize.SIZE_MEDIUM:
			return 'w-[35rem]';
		case ElementSize.SIZE_LARGE:
			return 'w-[45rem]';
		default:
			return 'w-auto';
	}
});
</script>
