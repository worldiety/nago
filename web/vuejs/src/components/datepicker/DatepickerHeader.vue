<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<!-- Datepicker header -->
	<div>
		<div class="flex justify-between items-center gap-x-2">
			<p class="truncate">{{ label ?? '' }}</p>
			<div
				ref="closeButton"
				class="flex justify-center items-center cursor-pointer rounded-full size-8 shrink-0 grow-0"
				tabindex="0"
				@click="$emit('close')"
				@keydown.enter="$emit('close')"
			>
				<Close class="h-4" />
			</div>
		</div>
		<div class="border-b border-b-primary mt-2"></div>
	</div>
</template>

<script setup lang="ts">
import { useTemplateRef } from 'vue';
import Close from '@/assets/svg/closeBold.svg';

defineProps<{
	label?: string;
}>();

defineEmits<{
	(e: 'close'): void;
}>();

const closeButton = useTemplateRef('closeButton');

defineExpose({ closeButton });
</script>
