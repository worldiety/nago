<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div class="ui-flow-chart" :style="frameStyles">
		<FlowChartActions
			v-if="flowChart && ui.toolbar && ui.toolbar.actions"
			:toolbar="ui.toolbar"
			:orientation="ui.orientation ?? OrientationValues.Horizontal"
			:flow-chart="flowChart"
			@action:auto-layout="autoLayout"
		/>
		<VueFlow
			ref="flowChart"
			class="ui-flow-chart__canvas"
			:class="{ dark: themeManager.getActiveThemeKey() === ThemeKey.DARK }"
			:nodes="nodes"
			:edges="edges"
			:nodes-draggable="!!ui.nodesDraggable"
			:nodes-connectable="!!ui.nodesConnectable"
			:edges-updatable="!!ui.edgesEditable"
			:elements-selectable="!!ui.elementsSelectable"
			:select-nodes-on-drag="false"
			:min-zoom="ui.minZoom || 0.2"
			:max-zoom="ui.maxZoom || 2"
			:default-viewport="{ zoom: 1 }"
			:fit-view-on-init="nodes.length > 0"
			@nodes-change="onNodesUpdate"
			@edges-change="onEdgesUpdate"
			@connect="onNodesConnect"
			@node-click="onNodeClick"
			@edge-click="onEdgeClick"
			@pane-click="onPaneClick"
		>
			<Background
				v-if="ui.background"
				:pattern-color="
					ui.background.gridColor
						? colorValue(ui.background.gridColor)
						: themeManager.getActiveThemeKey() === ThemeKey.DARK
							? '#aaa6'
							: '#aaa'
				"
				:gap="ui.background.gridGap"
				:variant="
					ui.background.gridStyle === FlowChartBackgroundGridStyleValues.FlowChartBackgroundGridStyleLines
						? 'lines'
						: 'dots'
				"
			/>
			<template #node-custom="node">
				<FlowChartCustomNode
					:node="node.data as FlowChartNode"
					:orientation="ui.orientation"
					:custom-contents="ui.customContents"
				/>
			</template>
		</VueFlow>
	</div>
</template>
<script lang="ts" setup>
import { computed, nextTick, onUnmounted, ref, watch } from 'vue';
import FlowChartCustomNode from '@/components/flowchart/FlowChartCustomNode.vue';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Background } from '@vue-flow/background';
import type { Connection, EdgeChange, EdgeMouseEvent, NodeChange, NodeMouseEvent, Styles } from '@vue-flow/core';
import { type Edge, type EdgeMarkerType, MarkerType, type Node, Position, VueFlow, useVueFlow } from '@vue-flow/core';
import {
	CallRequested,
	FlowChart,
	FlowChartAutoLayout,
	FlowChartNode,
	FlowChartNodes,
	NagoEvent,
	OrientationValues,
} from '@/shared/proto/nprotoc_gen';
import { FlowChartBackgroundGridStyleValues } from '@/shared/proto/nprotoc_gen';
import {
	FlowChartEdge,
	FlowChartEdgeMarkerValues,
	FlowChartEdgeStyleValues,
	FlowChartEdges,
	FlowChartModel,
	FlowChartNodeTypeValues,
	FlowChartPoint,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';
import { ThemeKey, useThemeManager } from '@/shared/themeManager';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';
import FlowChartActions from '@/components/flowchart/FlowChartActions.vue';
import ConnectionHandler from '@/shared/network/connectionHandler';
import { useLayout } from '@/components/flowchart/useLayout';

const props = defineProps<{
	ui: FlowChart;
}>();

const themeManager = useThemeManager();
const serviceAdapter = useServiceAdapter();
const { project } = useVueFlow();
let debounceTimer: number = 0;

const { layout } = useLayout();

const flowChart = ref<InstanceType<typeof VueFlow>>();

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);
	styles.push('position:relative');
	styles.push('overflow:hidden');

	if (!props.ui.frame?.height && !props.ui.frame?.minHeight) {
		styles.push('min-height:16rem');
	}

	if (props.ui.background?.color) {
		styles.push(`background-color:${colorValue(props.ui.background?.color)}`);
	}

	return styles.join(';');
});

const nodes = ref<Node[]>([]);
const edges = ref<Edge[]>([]);

function onUpdateNodes(updated: Node[]): void {
	nodes.value = updated;
	nextTick(() => {
		if (flowChart.value) flowChart.value.fitView();
	});

	const newValue = new FlowChartModel(
		new FlowChartNodes(updated.map((u) => u.data as FlowChartNode)),
		props.ui.value?.edges
	);
	debouncedInputModel(newValue);
}

function setNodesAndEdges() {
	setNodes();
	setEdges();
}

function setNodes() {
	props.ui.value?.nodes?.value.forEach((node, index) => {
		const id = node.id?.trim() || `${index}`;

		const newNode: Node = {
			id: id,
			type: 'custom',
			data: node,
			position: {
				x: node.position?.x ?? index * 180,
				y: node.position?.y ?? 0,
			},
			sourcePosition: mapSourcePosition(node),
			targetPosition: mapTargetPosition(node),
			focusable: true,
			class: nodeClass(node),
		};

		const idx = nodes.value.findIndex((e) => e.id === newNode.id);
		if (idx < 0) nodes.value.push(newNode);
		else nodes.value[idx] = newNode;
	});

	nodes.value = nodes.value.filter((node) => props.ui.value?.nodes?.value.some((n) => n.id === node.id));
}

function setEdges() {
	const availableNodeIds = new Set(nodes.value.map((node) => node.id));

	const newEdges: FlowChartEdge[] =
		props.ui.value?.edges?.value.filter((edge) => {
			const source = edge.sourceNodeId?.trim();
			const target = edge.targetNodeId?.trim();
			return source && target && availableNodeIds.has(source) && availableNodeIds.has(target);
		}) || [];

	edges.value = newEdges.map((edge) => {
		const source = edge.sourceNodeId?.trim();
		const target = edge.targetNodeId?.trim();
		const edgeColor = edge.color ? colorValue(edge.color) : undefined;

		return {
			id: edge.id?.trim() || `${source}-${target}-${randomStr(8)}`,
			data: edge,
			source: source as string,
			target: target as string,
			label: edge.label?.trim(),
			animated: edge.animated ?? false,
			focusable: true,
			markerStart: mapMarker(edge.markerStart, edgeColor),
			markerEnd: mapMarker(edge.markerEnd, edgeColor),
			style: edgeStyle(edge, edgeColor),
			labelShowBg: !!edge.label,
			labelBgPadding: [4, 6],
			labelBgBorderRadius: 6,
			labelBgStyle: props.ui.background?.color ? { fill: colorValue(props.ui.background.color) } : undefined,
			labelStyle: edgeColor ? { fill: edgeColor } : undefined,
		};
	});
}

function mapSourcePosition(node: FlowChartNode): Position | undefined {
	if (node.type === FlowChartNodeTypeValues.FlowChartNodeTypeEnd) {
		return undefined;
	}

	return Position.Right;
}

function mapTargetPosition(node: FlowChartNode): Position | undefined {
	if (node.type === FlowChartNodeTypeValues.FlowChartNodeTypeStart) {
		return undefined;
	}

	return Position.Left;
}

function nodeClass(node: FlowChartNode): string[] {
	const classes = ['ui-flow-chart__node'];

	if (node.type === FlowChartNodeTypeValues.FlowChartNodeTypeStart) {
		classes.push('ui-flow-chart__node--start');
	}

	if (node.type === FlowChartNodeTypeValues.FlowChartNodeTypeEnd) {
		classes.push('ui-flow-chart__node--end');
	}

	return classes;
}

function edgeStyle(edge: FlowChartEdge, edgeColor?: string): Styles {
	const style: Styles = {};

	if (edgeColor) {
		style.stroke = edgeColor;
	}

	if (edge.width !== undefined) {
		style.strokeWidth = edge.width;
	}

	switch (edge.style) {
		case FlowChartEdgeStyleValues.FlowChartEdgeStyleDashed:
			style.strokeDasharray = `${(edge.width ?? 1) * 6} ${(edge.width ?? 1) * 4}`;
			break;
		case FlowChartEdgeStyleValues.FlowChartEdgeStyleDotted:
			style.strokeDasharray = `${edge.width ?? 1} ${(edge.width ?? 1) * 4}`;
			break;
	}

	return style;
}

function mapMarker(marker: number | undefined, color?: string): EdgeMarkerType | undefined {
	switch (marker) {
		case FlowChartEdgeMarkerValues.FlowChartEdgeMarkerArrow:
			return { type: MarkerType.Arrow, color };
		default:
			return undefined;
	}
}

function onNodesUpdate(changes: NodeChange[]) {
	const valueCopy = new FlowChartModel(props.ui.value?.nodes, props.ui.value?.edges);

	changes.forEach((change) => {
		if (!valueCopy.nodes) return;

		switch (change.type) {
			case 'position':
				if (!change.position) return;
				valueCopy.nodes.value.forEach((node, index) => {
					if (!valueCopy.nodes) return;
					if (node.id === change.id) {
						valueCopy.nodes.value[index].position = new FlowChartPoint(
							change.position.x,
							change.position.y
						);
						debouncedInputModel(valueCopy);
					}
				});
				break;
			case 'remove':
				valueCopy.nodes.value = valueCopy.nodes?.value.filter((node) => node.id !== change.id);
				debouncedInputModel(valueCopy);
				onSelectionChange();
				break;
		}
	});
}

function onEdgesUpdate(changes: EdgeChange[]) {
	const valueCopy = new FlowChartModel(props.ui.value?.nodes, props.ui.value?.edges);

	changes.forEach((change) => {
		if (!valueCopy.edges) return;

		switch (change.type) {
			case 'remove':
				valueCopy.edges.value = valueCopy.edges?.value.filter((edge) => edge.id !== change.id);
				debouncedInputModel(valueCopy);
				onSelectionChange();
				break;
		}
	});
}

function onSelectionChange(): void {
	if (!flowChart.value) return;

	inputAction({
		selectedNodes: flowChart.value.getSelectedNodes.map((n) => n.id),
		selectedEdges: flowChart.value.getSelectedEdges.map((e) => e.id),
	});
}

function onNodesConnect(connection: Connection) {
	const valueCopy = new FlowChartModel(props.ui.value?.nodes, props.ui.value?.edges);
	const newFlowChartEdge = new FlowChartEdge(
		`edge-${connection.source}-${connection.target}-${randomStr(8)}`,
		connection.source,
		connection.target
	);
	if (!valueCopy.edges) valueCopy.edges = new FlowChartEdges();
	valueCopy.edges.value.push(newFlowChartEdge);
	debouncedInputModel(valueCopy);
}

function onNodeClick(e: NodeMouseEvent) {
	if (!flowChart.value) return;

	const event = e.event as PointerEvent;
	const node = e.node;

	const viewPos = {
		x: event.clientX,
		y: event.clientY,
	};
	const panePos = getPanePosFromClientPos(viewPos);

	inputAction({
		node: node.data,
		paneX: panePos.x,
		paneY: panePos.y,
		viewX: viewPos.x,
		viewY: viewPos.y,
		selectedNodes: flowChart.value.getSelectedNodes.map((n) => n.id),
		selectedEdges: flowChart.value.getSelectedEdges.map((e) => e.id),
	});
}

function onEdgeClick(e: EdgeMouseEvent) {
	if (!flowChart.value) return;

	const event = e.event as PointerEvent;
	const edge = e.edge;

	const viewPos = {
		x: event.clientX,
		y: event.clientY,
	};
	const panePos = getPanePosFromClientPos(viewPos);

	inputAction({
		edge: edge.data,
		paneX: panePos.x,
		paneY: panePos.y,
		viewX: viewPos.x,
		viewY: viewPos.y,
		selectedNodes: flowChart.value.getSelectedNodes.map((n) => n.id),
		selectedEdges: flowChart.value.getSelectedEdges.map((e) => e.id),
	});
}

function onPaneClick(e: MouseEvent) {
	nextTick(() => {
		if (!flowChart.value) return;

		const viewPos = {
			x: e.clientX,
			y: e.clientY,
		};
		const panePos = getPanePosFromClientPos(viewPos);

		inputAction({
			paneX: panePos.x,
			paneY: panePos.y,
			viewX: viewPos.x,
			viewY: viewPos.y,
			selectedNodes: flowChart.value.getSelectedNodes.map((n) => n.id),
			selectedEdges: flowChart.value.getSelectedEdges.map((e) => e.id),
		});
	});
}

function getPanePosFromClientPos(clientPos: { x: number; y: number }): { x: number; y: number } {
	if (!flowChart.value) return { x: 0, y: 0 };

	const flowChartBounds = flowChart.value.$el.getBoundingClientRect();
	clientPos.x -= flowChartBounds.left;
	clientPos.y -= flowChartBounds.top;

	return project(clientPos);
}

function inputAction(action: object) {
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.actionValue, 0, nextRID(), JSON.stringify(action)));
}

function debouncedInputModel(value: FlowChartModel) {
	const debounceTime = 100; // ms

	clearTimeout(debounceTimer);
	debounceTimer = window.setTimeout(() => {
		if (JSON.stringify(value) == JSON.stringify(props.ui)) return;
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(
				props.ui.inputValue,
				0,
				nextRID(),
				JSON.stringify({ nodes: value.nodes?.value, edges: value.edges?.value })
			)
		);
	}, debounceTime);
}

function autoLayout() {
	if (!flowChart.value) return;
	const nodes = layout(flowChart.value, props.ui.orientation ?? OrientationValues.Horizontal);
	onUpdateNodes(nodes);
}

function onNagoEvent(invoke: NagoEvent): void {
	if (!(invoke instanceof CallRequested)) return;

	if (invoke.call instanceof FlowChartAutoLayout) autoLayout();
}

function init() {
	setNodesAndEdges();
	ConnectionHandler.addEventListener(onNagoEvent);
}

watch(() => props.ui.value, setNodesAndEdges, { deep: true });
onUnmounted(() => ConnectionHandler.removeEventListener(onNagoEvent));

init();
</script>
<style scoped>
.ui-flow-chart__canvas {
	@apply min-h-[inherit];
}
</style>
