<script lang="ts" setup>
import { ref } from 'vue';
import type { CustomError } from '@/composables/errorhandling';
import UiGeneric from "@/components/UiGeneric.vue";
import type {LiveDialog} from "@/shared/model/liveDialog";
import type {LivePage} from "@/shared/model/livePage";
import Alert from '@/assets/svg/alert.svg'

const props = defineProps<{
	error: CustomError;
	ui: LiveDialog;
	page: LivePage;
}>();

const showAdditionalInformation = ref<boolean>(false);

function toggleErrorInfo(): void {
	showAdditionalInformation.value = !showAdditionalInformation.value;
}

function reload(): void {
	window.location.reload()
}
</script>

<template>
	<div class="relative z-50" aria-labelledby="modal-title" role="dialog" aria-modal="true">
		<div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
		<div class="fixed inset-0 z-10 w-screen overflow-y-auto">
			<div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
				<div class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg">
					<div class="bg-white px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
						<div class="sm:flex sm:items-start">
							<div class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
								<Alert/>
							</div>
							<div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
								<div class="mt-2">
									<p>{{ props.error.message }}</p>
									<button class="border border-black bg-white p-0.5 text-sm rounded-md" @click="toggleErrorInfo">Mehr Informationen</button>
									<p v-if="showAdditionalInformation">{{ props.error.additionalInformation }}</p>
								</div>
							</div>
						</div>
					</div>
					<div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
						<button type="button" class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto" @click="reload">Neu laden</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
