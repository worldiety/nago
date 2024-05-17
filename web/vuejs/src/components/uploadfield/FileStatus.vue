<template>
	<div class="flex justify-between items-center gap-x-4 text-black dark:text-white bg-disabled-background bg-opacity-15 rounded-lg dark:bg-opacity-5 w-full py-2 pl-4 pr-1">
		<div class="relative">
			<FileIcon class="h-4 text-disabled-text" />
			<div v-if="fileUpload.finished" class="absolute -bottom-1 -right-1 rounded-full bg-success h-3 p-0.5">
				<CheckIcon class="text-white dark:text-darkmode-gray h-full" />
			</div>
		</div>
		<div class="flex flex-col justify-between items-start gap-y-1 grow">
			<div class="flex justify-between items-center gap-x-2 w-full">
				<p>{{ fileUpload.file.name }}</p>
				<p class="grow text-sm text-disabled-text leading-none border-l border-l-disabled-text pl-2">{{ fileSizeFormatted }}</p>
				<p class="text-sm text-disabled-text">{{ informationText }}</p>
			</div>
			<progress v-if="progressBarVisible" :max="fileUpload.bytesTotal ?? 0" :value="fileUpload.bytesUploaded ?? 0" class="w-full"></progress>
		</div>
		<div
			v-if="!fileUpload.finished"
			class="cursor-pointer hover:bg-disabled-background hover:bg-opacity-35 active:bg-opacity-45 dark:hover:bg-opacity-10 dark:active:bg-opacity-15 rounded-full h-10 p-3"
			tabindex="0"
		>
			<CloseIcon class="text-disabled-text h-full" />
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
import { activeLocale } from '@/i18n';

const props = defineProps<{
	fileUpload: FileUpload;
}>();

const { t } = useI18n();

const progressBarVisible = computed((): boolean => {
	return !props.fileUpload.finished && props.fileUpload.bytesUploaded !== null && props.fileUpload.bytesTotal !== null;
});

const fileSizeFormatted = computed((): string => {
	if (props.fileUpload.bytesTotal === null) {
		return t('uploadField.unknown');
	}
	const fileSizeMegabytes = props.fileUpload.bytesTotal / 1000000;
	return `${t('uploadField.size')}: ${localizeNumber(fileSizeMegabytes, { maximumFractionDigits: 2 })} MB`;
});

const progressFormatted = computed((): string => {
	if (props.fileUpload.bytesUploaded === null || props.fileUpload.bytesTotal === null) {
		return '';
	}
	const progressPercentage = props.fileUpload.bytesUploaded / props.fileUpload.bytesTotal * 100;
	return `${localizeNumber(progressPercentage, { maximumFractionDigits: 0 })}%`;
});

const informationText = computed((): string => {
	if (!props.fileUpload.finished) {
		return progressFormatted.value;
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
</script>
