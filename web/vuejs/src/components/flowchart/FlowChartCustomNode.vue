<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div class="custom-node" :style="styles">
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
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { Handle, Position, type Styles } from '@vue-flow/core';
import type { Component, FlowChartCustomContents, FlowChartNode } from '@/shared/proto/nprotoc_gen';
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

const styles = computed<Styles>(() => {
	const style: Styles = {};

	if (props.node.backgroundColor) {
		style.backgroundColor = colorValue(props.node.backgroundColor);
	}

	Object.assign(style, cssDeclarationsToStyle(borderCSS(props.node.border)));

	return style;
});

function cssDeclarationsToStyle(declarations: string[]): Styles {
	const style: Record<string, string> = {};

	for (const declaration of declarations) {
		const separatorIndex = declaration.indexOf(':');
		if (separatorIndex < 0) {
			continue;
		}

		const property = declaration.slice(0, separatorIndex).trim();
		const value = declaration.slice(separatorIndex + 1).trim();
		if (!property || !value) {
			continue;
		}

		style[property] = value;
	}

	return style as Styles;
}
</script>
