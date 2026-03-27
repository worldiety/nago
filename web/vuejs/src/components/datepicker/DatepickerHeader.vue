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
			<button
				ref="closeButton"
				class="button-tertiary square size-auto -mr-2.5 !p-2.5 text-ST0 focus:text-I0 hover:opacity-90"
				tabindex="0"
				aria-label="Dialog schließen"
				@click="$emit('close')"
				@keydown.enter="$emit('close')"
			>
				<Close class="h-3.5" aria-hidden="true" />
			</button>
		</div>
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
