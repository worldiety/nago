<script lang="ts" setup>
import {computed, markRaw} from 'vue';
import {fetchUpload} from "@/api/upload/uploadRepository";
import {ApplicationError, useErrorHandling} from "@/composables/errorhandling";
import UiErrorMessage from "@/components/UiErrorMessage.vue";
import {FileField} from "@/shared/protocol/ora/fileField";

const errorHandler = useErrorHandling();

const props = defineProps<{
	ui: FileField;
}>();

function isErr(): boolean {
	return props.ui.error.v != '';
}

const labelClass = computed<string>(() => {
	if (props.ui.disabled.v && isErr()) {
		return 'text-red-900 dark:text-red-700';
	}

	if (isErr()) {
		return 'text-red-700 dark:text-red-500';
	}

	return 'text-gray-900 dark:text-white';
});

const inputClass = computed<string>(() => {
	if (props.ui.disabled.v) {
		return 'bg-gray-100 border border-gray-200 text-gray-600 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 cursor-not-allowed dark:bg-gray-800 dark:border-gray-600 dark:placeholder-gray-400 dark:text-gray-400 dark:focus:ring-blue-500 dark:focus:border-blue-500';
	}

	if (isErr()) {
		return 'bg-red-50 border border-red-500 text-red-900 placeholder-red-700 text-sm rounded-lg focus:ring-red-500 dark:bg-gray-700 focus:border-red-500 block w-full p-2.5 dark:text-red-500 dark:placeholder-red-500 dark:border-red-500';
	}

	return 'bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500';
});


async function fileInputChanged(e: Event):Promise<void> {
	const item = e.target as HTMLInputElement;
	if (!item.files) {
		return
	}
	const files: File[] = []
	for (let i = 0; i < item.files.length; i++) {
		files.push(item.files[i])
	}
	try {
		await fetchUpload(files, "???", props.ui.uploadToken.v) // todo backend must resolve page/scope whatever by token itself
	} catch (e: ApplicationError) {
		errorHandler.handleError(e)
	}
}



</script>

<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"> </UiErrorMessage>
	</div>
	<div v-else class="flex w-full items-center justify-center">
		<label
			:for="props.ui.id.toString()"
			class="dark:hover:bg-bray-800 flex h-64 w-full cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-700 dark:hover:border-gray-500 dark:hover:bg-gray-600"
		>
			<div class="flex flex-col items-center justify-center pb-6 pt-5">
				<svg
					class="mb-4 h-8 w-8 text-gray-500 dark:text-gray-400"
					aria-hidden="true"
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 20 16"
				>
					<path
						stroke="currentColor"
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"
					/>
				</svg>
				<p class="mb-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.v }}</p>
				<p class="text-xs text-gray-500 dark:text-gray-400">{{ props.ui.label.v }}</p>
			</div>
			<input
				@change="fileInputChanged"
				:disabled="props.ui.disabled.v"
				:id="props.ui.id.toString()"
				type="file"
				class="hidden"
				:multiple="props.ui.multiple.v"
				:accept="props.ui.filter.v"
			/>
			<p v-if="isErr()" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.v }}</p>
		</label>
	</div>
</template>
