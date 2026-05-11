<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div
		class="custom-node"
		:class="{
			'style-default': !node.style || node.style === FlowChartNodeStyleValues.FlowChartNodeStyleDefault,
			'style-none': node.style === FlowChartNodeStyleValues.FlowChartNodeStyleNone,
		}"
	>
		<Handle
			v-if="!node.type || node.type === FlowChartNodeTypeValues.FlowChartNodeTypeEnd"
			type="target"
			:position="orientation === OrientationValues.Vertical ? Position.Top : Position.Left"
		/>
		<UiGeneric v-if="customContent" :ui="customContent" />
		<div v-else class="label">{{ node.label }}</div>
		<Handle
			v-if="!node.type || node.type === FlowChartNodeTypeValues.FlowChartNodeTypeStart"
			type="source"
			:position="orientation === OrientationValues.Vertical ? Position.Bottom : Position.Right"
		/>
	</div>
</template>
<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { Handle, Position } from '@vue-flow/core';
import {
	Component,
	FlowChartCustomContents,
	FlowChartNode,
	FlowChartNodeStyleValues,
} from '@/shared/proto/nprotoc_gen';
import { FlowChartNodeTypeValues } from '@/shared/proto/nprotoc_gen';
import { OrientationValues } from '@/shared/proto/nprotoc_gen';

interface Props {
	node: FlowChartNode;
	orientation?: OrientationValues;
	customContents?: FlowChartCustomContents;
}

const props = defineProps<Props>();

const customContent = computed<Component | undefined>(() => {
	return props.customContents?.value.find((c) => c.nodeId === props.node.id)?.content;
});
</script>
