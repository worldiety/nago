<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"></UiErrorMessage>
	</div>
	<div v-else class="flex flex-col items-start justify-center gap-y-2 w-full">
		<p v-if="isErr()" class="text-sm text-error text-end w-full">{{ props.ui.error.v }}</p>
		<div class="dark:hover:bg-bray-800 flex h-64 w-full cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-700 dark:hover:border-gray-500 dark:hover:bg-gray-600">
			<div class="flex flex-col items-center justify-center pb-6 pt-5">
				<UploadIcon class="mb-4 h-8 w-8 text-gray-500 dark:text-gray-400" />
				<p class="mb-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.label.v }}</p>
			</div>
			<input
				:disabled="props.ui.disabled.v"
				type="file"
				class="hidden"
				:multiple="props.ui.multiple.v"
				:accept="props.ui.filter.v"
				@change="fileInputChanged"
			/>
		</div>
		<div class="flex justify-between items-center text-sm w-full">
			<p class="text-gray-500 dark:text-gray-400">{{ props.ui.hintLeft.v }}</p>
			<p class="text-gray-500 dark:text-gray-400">{{ props.ui.hintRight.v }}</p>
		</div>
	</div>
</template>

<script setup lang="ts">
import { fetchUpload } from "@/api/upload/uploadRepository";
import { ApplicationError, useErrorHandling } from "@/composables/errorhandling";
import UiErrorMessage from "@/components/UiErrorMessage.vue";
import type { FileField } from "@/shared/protocol/ora/fileField";
import UploadIcon from '@/assets/svg/upload.svg';

const props = defineProps<{
	ui: FileField;
}>();

const errorHandler = useErrorHandling();

function isErr(): boolean {
	return props.ui.error.v != '';
}

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
