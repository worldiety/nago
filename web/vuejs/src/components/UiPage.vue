<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {Page} from "@/shared/protocol/ora/page";
import GenericUi from "@/components/UiGeneric.vue";



const props = defineProps<{
	ui: Page;
}>();

</script>

<template>
	<div class="relative z-50" aria-labelledby="modal-title" role="dialog" aria-modal="true">


		<div v-for="modal in ui.modals.v">
			<div class="fixed inset-0 z-50 bg-gray-700 bg-opacity-75 transition-opacity"></div>

			<div class="fixed inset-0 z-50 w-screen overflow-y-auto">
				<div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
					<div class="relative transform rounded-lg sm:my-8 sm:w-full sm:max-w-lg">
						<generic-ui   :ui="modal"/>
					</div>
				</div>
			</div>
		</div>
	</div>

	<div class="grid grid-cols-1 gap-y-3">
		<ui-generic v-if="props.ui.body.v" :ui="props.ui.body.v"  />
	</div>
</template>
