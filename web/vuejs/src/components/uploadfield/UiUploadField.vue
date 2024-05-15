<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"></UiErrorMessage>
	</div>
	<div v-else class="flex flex-col items-start justify-center gap-y-2 w-full">
		<p v-if="isErr()" class="text-sm text-error text-end w-full">{{ props.ui.error.v || errorMessage }}</p>
		<div
			class="dark:hover:bg-bray-800 flex h-64 w-full cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-disabled-text bg-disabled-background bg-opacity-15 hover:bg-opacity-25 dark:bg-opacity-5 dark:hover:bg-opacity-10"
			tabindex="0"
			@click="showUploadDialog"
			@keydown.enter="showUploadDialog"
		>
			<div class="flex flex-col items-center justify-center pb-6 pt-5">
				<UploadIcon class="mb-4 h-8 w-8 text-gray-500 dark:text-gray-400" />
				<p class="mb-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.label.v }}</p>
			</div>
			<input
				ref="fileInput"
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

		<!-- File statuses -->
		<template v-if="fileUploads && bytesUploaded !== null && bytesTotal !== null">
			<FileStatus
				v-for="(fileUpload, index) in fileUploads"
				:key="index"
				:file-upload="fileUpload"
			/>
		</template>
	</div>
</template>

<script setup lang="ts">
import { fetchUpload } from "@/api/upload/uploadRepository";
import { ApplicationError, useErrorHandling } from "@/composables/errorhandling";
import UiErrorMessage from "@/components/UiErrorMessage.vue";
import type { FileField } from "@/shared/protocol/ora/fileField";
import UploadIcon from '@/assets/svg/upload.svg';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import FileStatus from '@/components/uploadfield/FileStatus.vue';
import type FileUpload from '@/components/uploadfield/fileUpload';
import { v4 as uuidv4 } from 'uuid';
import {useServiceAdapter} from "@/composables/serviceAdapter";

const props = defineProps<{
	ui: FileField;
}>();

const errorHandler = useErrorHandling();
const { t } = useI18n();
const fileInput = ref<HTMLElement|undefined>();
const errorMessage = ref<string|null>(null);
const fileUploads = ref<FileUpload[]|null>(null);
const bytesUploaded = ref<number|null>(null);
const bytesTotal = ref<number|null>(null);

const serviceAdapter = useServiceAdapter();

function showUploadDialog(): void {
	fileInput.value?.click();
}

function isErr(): boolean {
	return props.ui.error.v != '' || errorMessage.value !== null;
}

async function fileInputChanged(e: Event):Promise<void> {
	const item = e.target as HTMLInputElement;
	if (!item.files) {
		return;
	}
	const filesToUpload: File[] = []
	for (let i = 0; i < item.files.length; i++) {
		filesToUpload.push(item.files[i])
	}
	if (!filesValid(filesToUpload)) {
		errorHandler.handleError({
			errorCode: '003',
			message: t('customErrorcodes.003.errorMessage', [`${props.ui.maxBytes.v / 1000000} MB`]),
		});
		return;
	}

	fileUploads.value = filesToUpload.map((file) => ({
		uploadId: uuidv4(),
		file,
		bytesUploaded: null,
		bytesTotal: null,
		finished: false,
	}));
	const promises = fileUploads.value.map((fileUpload) => {
		return fetchUpload(
			fileUpload.file,
			fileUpload.uploadId,
			props.ui.id,
			serviceAdapter.getScopeID(),
			uploadProgressCallback,
		)
	});
	try {
		await Promise.all(promises);
	} catch (e: ApplicationError) {
		errorHandler.handleError(e);
	}
}

function filesValid(files: File[]): boolean {
	if (!props.ui.maxBytes.v && props.ui.maxBytes.v !== 0) {
		return true;
	}
	return files.every((file) => file.size <= props.ui.maxBytes.v);
}

function uploadProgressCallback(uploadId: string, progress: number, total: number): void {
	if (!fileUploads.value) {
		return;
	}
	// TODO: Check, if still reactive as soon as upload is working again
	fileUploads.value = fileUploads.value.map((fileUpload) => {
		if (fileUpload.uploadId === uploadId) {
			return {
				...fileUpload,
				bytesUploaded: progress,
				bytesTotal: total,
			};
		}
		return fileUpload;
	});
}
</script>
