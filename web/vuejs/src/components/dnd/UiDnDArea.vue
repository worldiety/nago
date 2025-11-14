<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useDnDState } from '@/components/dnd/useDnDState';
import { frameCSS } from '@/components/shared/frame';
import { bool2Str } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { DnDArea, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: DnDArea;
}>();

const emit = defineEmits<{
	(e: 'drop-success', id: string): void;
}>();

const isOver = ref(false);
const canDrop = ref(false);
const isDragging = ref(false);

const serviceAdapter = useServiceAdapter();

const { currentDragId, startDrag, endDrag } = useDnDState();

const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

function onDragStart(event: DragEvent) {
	if (!event.dataTransfer) {
		return;
	}

	isDragging.value = true;

	if (props.ui.id) {
		startDrag(props.ui.id);
		event.dataTransfer.setData('application/x-drag-id', props.ui.id);
	}

	event.dataTransfer.effectAllowed = 'move';
	event.dataTransfer.dropEffect = canDrop.value ? 'move' : 'none';
}

function onDragOver(event: DragEvent) {
	isOver.value = true;
	// const dragId = event.dataTransfer?.getData('application/x-drag-id');
	const dragId = currentDragId.value;
	if (props.ui.dnD?.droppableIDs) {
		canDrop.value = !!dragId && props.ui.dnD?.droppableIDs.value.includes(dragId);
	} else {
		canDrop.value = true;
	}

	if (canDrop.value) {
		event.dataTransfer!.dropEffect = 'move';
	} else {
		event.dataTransfer!.dropEffect = 'none';
	}

	document.body.style.cursor = canDrop.value ? 'grab' : 'not-allowed';
}

function onDragLeave() {
	isOver.value = false;
	canDrop.value = false;
	isDragging.value = false;
	document.body.style.cursor = 'default';
}

function onDrop(event: DragEvent) {
	//const dragId = event.dataTransfer?.getData('application/x-drag-id');
	const dragId = currentDragId.value;
	isOver.value = false;
	canDrop.value = false;
	isDragging.value = false;
	document.body.style.cursor = 'default';
	if (props.ui.dnD?.droppableIDs) {
		if (dragId && props.ui.dnD?.droppableIDs?.value.includes(dragId)) {
			emit('drop-success', dragId);
			serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.droppedId, 0, nextRID(), dragId));
		}
	} else if (dragId) {
		emit('drop-success', dragId);
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.droppedId, 0, nextRID(), dragId));
	}

	endDrag();
}
</script>

<template>
	<!-- dnd area -->
	<div
		@dragover.prevent="onDragOver"
		@drop.prevent="onDrop"
		@dragleave="onDragLeave"
		:style="styles"
		:draggable="ui.dnD?.canDrag"
		@dragstart="onDragStart"
		:id="props.ui.id"
		:class="{
			'border border-green-500 bg-green-50 cursor-move': isOver && canDrop,
			'border border-red-400 bg-red-50 border cursor-not-allowed': isOver && !canDrop,
		}"
	>
		<!--		<div v-if="isOver && !canDrop" class="absolute text-red-500 text-sm mt-2">
			❌ Nicht erlaubt
		</div>-->

		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</div>
</template>
