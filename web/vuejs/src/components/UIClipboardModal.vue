<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ClipboardIcon from '@/assets/svg/clipboard.svg';

const { t } = useI18n();

const props = defineProps<{
	modelValue: string;
}>();
const emit = defineEmits(['update:modelValue']);

const copy_text = () => {
	navigator.clipboard.writeText(props.modelValue);
	close_modal();
};

const close_modal = () => {
	emit('update:modelValue', '');
};
</script>

<template>
	<div v-if="modelValue" class="copy-to-clipboard-container">
		<div class="flex justify-between align-center z-5 copy-to-clipboard-box">
			<div class="flex row py-4 items-center">
				<ClipboardIcon class="icon" />
				<p class="font-body-medium pr-4">{{ t('clipboardModal.info') }}</p>
			</div>
			<div>
				<button class="button-primary" @click="close_modal">
					<span>{{ t('clipboardModal.cancel') }}</span>
				</button>
				<button class="button-primary" @click="copy_text">
					<span>{{ t('clipboardModal.confirm') }}</span>
				</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
.copy-to-clipboard-container {
	display: flex;
	position: fixed;
	z-index: 1;
	left: 0;
	top: 0;
	width: 100%;
	height: 100%;
	justify-content: center;
	align-items: center;
	overflow: auto;
	background-color: rgba(0, 0, 0, 0.4);
}

.copy-to-clipboard-box {
	display: flex;
	position: fixed;
	flex-direction: column;
	background-color: #fefefe;
	color: black;
	width: 20%;
	max-width: 18em;
	min-width: 12em;
	box-sizing: border-box;
}

.copy-to-clipboard-box .button-primary {
	font-size: 1rem;
	box-sizing: border-box;
}

.copy-to-clipboard-box .icon {
	min-width: 32px;
	height: 32px;
	margin: 1.5em;
	padding: 0;
	align-items: center;
	justify-items: center;
}

.copy-to-clipboard-box button {
	background: transparent;
	justify-self: stretch;
	margin: 0;
	padding: 0;
	border-radius: 0;
	border-top: gray 1px solid;
	width: 50%;
}

.copy-to-clipboard-box button:first-of-type {
	border-right: gray 1px solid;
}

.copy-to-clipboard-box button:hover {
	background: rgba(0, 0, 0, 0.1);
}
</style>
