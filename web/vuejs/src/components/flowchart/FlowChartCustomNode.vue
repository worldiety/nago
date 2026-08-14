<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div
		ref="container"
		class="custom-node"
		:class="{
			'style-default': !node.style || node.style === FlowChartNodeStyleValues.FlowChartNodeStyleDefault,
			'style-none': node.style === FlowChartNodeStyleValues.FlowChartNodeStyleNone,
			'readonly': readonly,
		}"
	>
		<Handle
			v-if="!node.type || node.type === FlowChartNodeTypeValues.FlowChartNodeTypeEnd"
			type="target"
			:position="orientation === OrientationValues.Vertical ? Position.Top : Position.Left"
		/>
		<div>
			<UiGeneric v-if="customContent && customContent.content" :ui="customContent.content" />
			<div v-else class="label">{{ node.label }}</div>
			<div v-if="customContent && customContent.menu && zoom" :style="`transform: scale(${1 / zoom});`">
				<div ref="menu" :style="floatingStyles">
					<UiGeneric :ui="customContent.menu" />
				</div>
			</div>
		</div>
		<Handle
			v-if="!node.type || node.type === FlowChartNodeTypeValues.FlowChartNodeTypeStart"
			type="source"
			:position="orientation === OrientationValues.Vertical ? Position.Bottom : Position.Right"
		/>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { Handle, Position } from '@vue-flow/core';
import {
	FlowChartCustomContent,
	FlowChartCustomContents,
	FlowChartNode,
	FlowChartNodeStyleValues,
} from '@/shared/proto/nprotoc_gen';
import { FlowChartNodeTypeValues } from '@/shared/proto/nprotoc_gen';
import { OrientationValues } from '@/shared/proto/nprotoc_gen';
import { autoUpdate, offset, shift, useFloating } from '@floating-ui/vue';

interface Props {
	node: FlowChartNode;
	orientation?: OrientationValues;
	customContents?: FlowChartCustomContents;
	zoom?: number;
	readonly?: boolean;
}

const props = defineProps<Props>();

const container = ref();
const menu = ref();

const { floatingStyles } = useFloating(container, menu, {
	placement: 'top',
	strategy: 'fixed',
	whileElementsMounted: autoUpdate,
	middleware: [offset(8), shift({ crossAxis: true })],
});

const customContent = computed<FlowChartCustomContent | undefined>(() => {
	return props.customContents?.value.find((c) => c.nodeId === props.node.id);
});
</script>
