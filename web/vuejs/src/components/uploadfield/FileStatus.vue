<template>
	<div
		class="text-black dark:text-white bg-disabled-background bg-opacity-15 rounded-lg dark:bg-opacity-5 w-full py-2 pl-4"
		:class="inProgress || pending ? 'pr-1' : 'pr-4'"
	>
		<div class="flex justify-between items-center gap-x-4 h-full min-h-10">
			<div class="relative">
				<FileIcon class="h-4 text-disabled-text" />
				<div v-if="fileUpload.status === FileUploadStatus.SUCCESS" class="absolute -bottom-1 -right-1 rounded-full bg-success h-3 p-0.5">
					<CheckIcon class="text-white dark:text-darkmode-gray h-full" />
				</div>
				<div v-else-if="aborted || errorOccurred" class="absolute -bottom-1 -right-1 rounded-full bg-error h-3 p-[0.2rem]">
					<CloseIcon class="text-white dark:text-darkmode-gray h-full" />
				</div>
			</div>
			<div class="flex flex-col justify-between items-start gap-y-1 grow overflow-hidden">
				<div class="flex flex-wrap justify-between items-center gap-x-2 whitespace-nowrap w-full">
					<p class="truncate">{{ fileUpload.file.name }}</p>
					<p class="grow text-sm text-disabled-text leading-none border-l border-l-disabled-text pl-2">{{ fileSizeFormatted }}</p>
					<p class="text-sm text-disabled-text">{{ informationText }}</p>
				</div>
				<progress v-if="inProgress" :max="fileUpload.bytesTotal ?? 0" :value="fileUpload.bytesUploaded ?? 0" class="duration-200 w-full"></progress>
			</div>
			<div
				v-if="inProgress"
				class="text-disabled-text cursor-pointer hover:bg-ora-orange hover:bg-opacity-15 hover:text-ora-orange active:bg-opacity-20 rounded-full size-10 p-3"
				tabindex="0"
				@click="abortUpload"
				@keydown.enter="abortUpload"
			>
				<CloseIcon class="h-full" />
			</div>
			<div v-else-if="pending" class="text-disabled-text size-10 p-2">
				<LoadingAnimation class="h-full" />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import FileIcon from '@/assets/svg/file.svg';
import CheckIcon from '@/assets/svg/check.svg';
import CloseIcon from '@/assets/svg/closeBold.svg';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { localizeNumber } from '@/shared/localization';
import type FileUpload from '@/components/uploadfield/fileUpload';
import { FileUploadStatus } from '@/components/uploadfield/fileUpload';
import { activeLocale } from '@/i18n';
import { useUploadRepository } from '@/api/upload/uploadRepository';
import LoadingAnimation from '@/components/shared/LoadingAnimation.vue';

const props = defineProps<{
	fileUpload: FileUpload;
}>();

const uploadRepository = useUploadRepository();
const { t } = useI18n();

const inProgress = computed((): boolean => {
	return props.fileUpload.status === FileUploadStatus.IN_PROGRESS
		&& props.fileUpload.bytesUploaded !== null
		&& props.fileUpload.bytesTotal !== null;
});

const pending = computed((): boolean => {
	return props.fileUpload.status === FileUploadStatus.PENDING;
});

const aborted = computed((): boolean => {
	return props.fileUpload.status === FileUploadStatus.ABORTED;
});

const errorOccurred = computed((): boolean => {
	return props.fileUpload.status === FileUploadStatus.ERROR;
});

const fileSizeFormatted = computed((): string => {
	if (props.fileUpload.bytesTotal === null) {
		return t('uploadField.unknown');
	}

	let fileSizeFormatted: number;
	let fileSizeUnit: string;
	if (props.fileUpload.bytesTotal < Math.pow(10, 3)) {
		fileSizeUnit = 'B';
		fileSizeFormatted = props.fileUpload.bytesTotal;
	} else if (props.fileUpload.bytesTotal < Math.pow(10, 6)) {
		fileSizeUnit = 'KB';
		fileSizeFormatted = props.fileUpload.bytesTotal / Math.pow(10, 3);
	} else if (props.fileUpload.bytesTotal < Math.pow(10, 9)) {
		fileSizeUnit = 'MB';
		fileSizeFormatted = props.fileUpload.bytesTotal / Math.pow(10, 6);
	} else {
		fileSizeUnit = 'GB';
		fileSizeFormatted = props.fileUpload.bytesTotal / Math.pow(10, 9);
	}
	return `${t('uploadField.size')}: ${localizeNumber(fileSizeFormatted, { maximumFractionDigits: 2 })} ${fileSizeUnit}`;
});

const progressFormatted = computed((): string => {
	if (props.fileUpload.bytesUploaded === null || props.fileUpload.bytesTotal === null) {
		return '';
	}
	const progressPercentage = props.fileUpload.bytesUploaded / props.fileUpload.bytesTotal * 100;
	return `${localizeNumber(progressPercentage, { maximumFractionDigits: 0 })}%`;
});

const informationText = computed((): string => {
	if (pending.value) {
		return t('uploadField.pending')
	}
	if (inProgress.value) {
		return progressFormatted.value;
	}
	if (aborted.value) {
		return t('uploadField.aborted');
	}
	if (errorOccurred.value) {
		return t('uploadField.errorOccurred', [props.fileUpload.statusCode ?? 'Unbekannt']);
	}
	const currentDate = new Date();
	const uploadDateString = currentDate.toLocaleString(activeLocale, {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit',
	});
	return `Hochgeladen am ${uploadDateString}`;
});

function abortUpload(): void {
	uploadRepository.abortUpload(props.fileUpload.uploadId);
}
</script>
