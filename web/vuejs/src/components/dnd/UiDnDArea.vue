<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useDndStore } from '@/components/dnd/useDndStore';
import { frameCSS } from '@/components/shared/frame';
import { randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { DnDArea } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: DnDArea;
}>();

const emit = defineEmits<{
	(e: 'drop-success', id: string): void;
}>();

const id = props.ui.id || randomStr(16);
const serviceAdapter = useServiceAdapter();
const store = useDndStore();

const isOver = ref(false);
const isDropArea = ref(!!props.ui.dnD?.canDrop);
const canDrop = ref(false);

const styles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);
	return styles.join(';');
});

function onDragStart(event: DragEvent) {
	if (!event.dataTransfer) return;

	store.startDrag(id);
	event.dataTransfer.setData('application/x-drag-id', id);
	event.dataTransfer.effectAllowed = 'all';
}

function onDragEnd(event: DragEvent) {
	if (!event.dataTransfer) return;

	store.endDrag();
}

function onDragOver(event: DragEvent) {
	isOver.value = true;
	event.dataTransfer!.dropEffect = canDrop.value ? 'copy' : 'none';
}

function onDragEnter(event: DragEvent) {
	// delay by a tick to allow 'drag leave' events to complete first
	nextTick(() => {
		if (!isDropArea.value || !event.dataTransfer) return;

		const dragId = store.currentDragId;
		if (props.ui.dnD?.canDrop && props.ui.dnD?.droppableIDs) {
			canDrop.value = !!dragId && props.ui.dnD?.droppableIDs.value.includes(dragId);
			store.enteredDropArea(canDrop.value);
		} else {
			canDrop.value = !!props.ui.dnD?.canDrop;
			store.enteredDropArea(canDrop.value);
		}
	});
}

function onDragLeave() {
	if (!isDropArea.value) return;
	isOver.value = false;
	canDrop.value = false;
	store.leftDropArea();
}

function onDrop() {
	const dragId = store.currentDragId;
	isOver.value = false;
	canDrop.value = false;
	if (props.ui.dnD?.droppableIDs) {
		if (dragId && props.ui.dnD?.droppableIDs?.value.includes(dragId)) {
			emit('drop-success', dragId);
			serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.droppedId, 0, nextRID(), dragId));
		}
	} else if (dragId) {
		emit('drop-success', dragId);
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.droppedId, 0, nextRID(), dragId));
	}

	store.endDrag();
}
</script>

<template>
	<!-- dnd area -->
	<div
		:id="id"
		:class="{
			'outline outline-1 -outline-offset-1': isDropArea && isOver,
			'outline-green-500 bg-green-50': isDropArea && isOver && canDrop,
			'outline-red-600 bg-red-50': isDropArea && isOver && !canDrop,
		}"
		:style="styles"
		:draggable="ui.dnD?.canDrag"
		@dragover.prevent="onDragOver"
		@drop.prevent="onDrop"
		@dragstart="onDragStart"
		@dragend="onDragEnd"
		@dragenter="onDragEnter"
		@dragleave="onDragLeave"
	>
		<!--		<div v-if="isOver && !canDrop" class="absolute text-red-500 text-sm mt-2">
			❌ Nicht erlaubt
		</div>-->

		<ui-generic
			v-for="ui in props.ui.children?.value"
			:ui="ui"
			:class="store.dragging ? '!pointer-events-none' : ''"
		/>
	</div>
</template>
