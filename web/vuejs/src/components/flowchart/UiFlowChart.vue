<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div ref="flowChartContainer" class="ui-flow-chart" :style="frameStyles" @keydown="onContainerKeyDown">
		<FlowChartActions
			v-if="flowChart && ui.toolbar && ui.toolbar.actions"
			:toolbar="ui.toolbar"
			:orientation="ui.orientation ?? OrientationValues.Horizontal"
			@action:auto-layout="autoLayout"
			@action:zoom-in="zoomIn"
			@action:zoom-out="zoomOut"
			@action:fit-view="fitView"
		/>
		<VueFlow
			ref="flowChart"
			class="ui-flow-chart__canvas"
			:class="{
				dark: themeManager.getActiveThemeKey() === ThemeKey.DARK,
				ungrabbable: ui.readOnly || !ui.nodesDraggable,
			}"
			:nodes="nodes"
			:edges="edges"
			:nodes-draggable="!ui.readOnly && !!ui.nodesDraggable"
			:nodes-connectable="!ui.readOnly && !!ui.nodesConnectable"
			:edges-updatable="!ui.readOnly && !!ui.edgesEditable"
			:elements-selectable="!!ui.elementsSelectable"
			:select-nodes-on-drag="false"
			:min-zoom="ui.minZoom || 0.2"
			:max-zoom="ui.maxZoom || 2"
			:default-viewport="{ zoom: 1 }"
			:fit-view-on-init="nodes.length > 0"
			:prevent-scrolling="!ui.movable"
			:pan-on-drag="!!ui.movable"
			:pan-on-scroll="!!ui.movable"
			:zoom-on-double-click="!!ui.movable"
			:zoom-on-scroll="!!ui.movable"
			:zoom-on-pinch="!!ui.movable"
			@move="enforceViewportLock"
			@pane-ready="onPaneReady"
			@nodes-change="onNodesUpdate"
			@edges-change="onEdgesUpdate"
			@connect="onNodesConnect"
			@node-click="onNodeClick"
			@node-context-menu="onNodeContextMenu"
			@edge-click="onEdgeClick"
			@edge-context-menu="onEdgeContextMenu"
			@pane-click="onPaneClick"
			@pane-context-menu.prevent="onPaneContextMenu"
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
					v-if="node.data instanceof FlowChartNode"
					:node="node.data"
					:orientation="ui.orientation"
					:custom-contents="ui.customContents"
					:zoom="currentZoom"
					:readonly="ui.readOnly"
				/>
				<FlowChartCustomMenu
					v-if="node.data instanceof FlowChartMenu"
					:content="node.data.content as Component"
					:zoom="currentZoom"
				/>
			</template>
		</VueFlow>

		<span id="aria-description-edge" class="aria-hidden">
			{{ $t('flowChart.edge.accessibilityDescription') }}
		</span>
		<span id="aria-description-node" class="aria-hidden">
			{{ $t('flowChart.node.accessibilityDescription') }}
		</span>
		<span class="aria-hidden" role="status" aria-live="polite" aria-atomic="true">
			{{ selectionAnnouncement }}
		</span>
	</div>
</template>
<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch, watchEffect } from 'vue';
import FlowChartCustomNode from '@/components/flowchart/FlowChartCustomNode.vue';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Background } from '@vue-flow/background';
import type { Connection, EdgeChange, EdgeMouseEvent, NodeChange, NodeMouseEvent, Styles } from '@vue-flow/core';
import {
	type Edge,
	type EdgeMarkerType,
	MarkerType,
	type Node,
	type ViewportTransform,
	Position,
	VueFlow,
	useVueFlow,
} from '@vue-flow/core';
import {
	CallRequested,
	Component,
	FlowChart,
	FlowChartAutoLayout,
	FlowChartMenu,
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
import FlowChartCustomMenu from '@/components/flowchart/FlowChartCustomMenu.vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
	ui: FlowChart;
}>();

const { t } = useI18n();
const themeManager = useThemeManager();
const serviceAdapter = useServiceAdapter();
const { project } = useVueFlow();
let debounceTimer: number = 0;

const { layout } = useLayout();

const flowChart = ref<InstanceType<typeof VueFlow>>();
const flowChartContainer = ref<HTMLElement>();

const nodes = ref<Node[]>([]);
const edges = ref<Edge[]>([]);

let resizeObserver: ResizeObserver | undefined;
const lockedViewport = ref<ViewportTransform>();

const currentZoom = computed<number | undefined>(() => {
	return flowChart.value?.getViewport().zoom;
});

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

function onPaneReady(): void {
	if (!props.ui.movable && flowChart.value) {
		lockedViewport.value = flowChart.value.getViewport();
	}
}

function enforceViewportLock(): void {
	if (props.ui.movable || !lockedViewport.value || !flowChart.value) return;

	const currentViewport = flowChart.value.getViewport();
	const viewport = lockedViewport.value;
	if (
		currentViewport.x !== viewport.x ||
		currentViewport.y !== viewport.y ||
		currentViewport.zoom !== viewport.zoom
	) {
		flowChart.value.setViewport(viewport, { duration: 0 });
	}
}

async function fitViewOnResize(): Promise<void> {
	if (!props.ui.readOnly || props.ui.movable || !flowChart.value) return;

	lockedViewport.value = undefined;
	await flowChart.value.fitView({ duration: 0 });

	if (!props.ui.movable && flowChart.value) {
		lockedViewport.value = flowChart.value.getViewport();
	}
}

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
		const customContent = props.ui.customContents?.value.find((cc) => cc.nodeId === node.id);
		const hasMenu = !!customContent?.menu;
		// "graphics-symbol" makes its children presentational, which would hide interactive custom content
		// from screen readers, so nodes with custom content are exposed as a group instead.
		const hasCustomContent = !!customContent?.content || hasMenu;

		const newNode: Node = {
			id: id,
			type: 'custom',
			data: node,
			ariaLabel: getNodeAccessibilityLabel(node),
			position: {
				x: node.position?.x ?? 0,
				y: node.position?.y ?? 0,
			},
			sourcePosition: mapSourcePosition(node),
			targetPosition: mapTargetPosition(node),
			focusable: true,
			class: nodeClass(node),
			zIndex: hasMenu ? 1 : undefined,
			domAttributes: {
				'aria-describedby': props.ui.elementsSelectable ? 'aria-description-node' : undefined,
				'role': hasCustomContent ? 'group' : 'graphics-symbol',
				'aria-roledescription': t('flowChart.node.accessibilityRoleDescription'),
				'tabindex': 0,
			},
		};

		const idx = nodes.value.findIndex((e) => e.id === newNode.id);
		if (idx < 0) nodes.value.push(newNode);
		else nodes.value[idx] = newNode;
	});

	nodes.value = nodes.value.filter((node) => props.ui.value?.nodes?.value.some((n) => n.id === node.id));

	if (props.ui.menu?.position && props.ui.menu.content) {
		const newNode: Node = {
			id: 'menu',
			type: 'custom',
			data: props.ui.menu,
			position: {
				x: props.ui.menu.position.x ?? 0,
				y: props.ui.menu.position.y ?? 0,
			},
			draggable: false,
			selectable: false,
			connectable: false,
			focusable: false,
			deletable: false,
			zIndex: 2,
		};

		nodes.value.push(newNode);
	}
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

		const edgeId = edge.id?.trim() || `${source}-${target}-${randomStr(8)}`;

		return {
			id: edgeId,
			data: edge,
			source: source as string,
			target: target as string,
			label: edge.label?.trim(),
			ariaLabel: getEdgeAccessibilityLabel(edge),
			animated: edge.animated ?? false,
			focusable: true,
			markerStart: mapMarker(edge.markerStart, edgeColor),
			markerEnd: mapMarker(edge.markerEnd, edgeColor),
			style: edgeStyle(edge, edgeColor),
			labelShowBg: !!edge.label,
			labelBgPadding: [6, 6],
			labelBgBorderRadius: 6,
			labelBgStyle: { fill: colorValue(edge.labelBgColor ?? props.ui.background?.color) },
			labelStyle: { fill: edge.labelTextColor ? colorValue(edge.labelTextColor) : edgeColor },
			domAttributes: {
				'aria-describedby': props.ui.elementsSelectable ? 'aria-description-edge' : undefined,
				'role': 'graphics-symbol',
				'aria-roledescription': t('flowChart.edge.accessibilityRoleDescription'),
				// Vue Flow sets "tabIndex" and "onKeyDown" on the edge's SVG element, which neither the browser
				// (case-sensitive SVG attributes) nor Vue (event "key-down") recognize, so we set them correctly here.
				'tabindex': 0,
				'onKeydown': (event: KeyboardEvent) => onEdgeKeyDown(event, edgeId),
			},
		};
	});
}

function getNodeAccessibilityLabel(node: FlowChartNode): string {
	const nodeLabel = node.label ?? node.id ?? '?';
	return t('flowChart.node.accessibilityLabel', { node: nodeLabel });
}

function getEdgeAccessibilityLabel(edge: FlowChartEdge): string {
	const sourceNode = nodes.value.find((n) => n.id === edge.sourceNodeId);
	const targetNode = nodes.value.find((n) => n.id === edge.targetNodeId);

	const sourceLabel = sourceNode?.data.label ?? sourceNode?.id ?? '?';
	const targetLabel = targetNode?.data.label ?? targetNode?.id ?? '?';

	const describing = t('flowChart.edge.accessibilityLabel', { source: sourceLabel, target: targetLabel });

	if (edge.label) return `${describing}: ${edge.label}`;
	return describing;
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
				restoreFocusAfterRemoval();
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
				restoreFocusAfterRemoval();
				break;
		}
	});
}

let focusRestorePending = false;

// Vue Flow blurs a node after deselecting it via Escape, which leaves keyboard and screen reader users without
// focus, so they could not reselect the node with Enter. The focus is moved back to the node after that blur.
function onContainerKeyDown(event: KeyboardEvent): void {
	if (event.key !== 'Escape') return;

	const target = event.target;
	if (!(target instanceof HTMLElement) || !target.classList.contains('vue-flow__node')) return;

	window.setTimeout(() => {
		if (target.isConnected && document.activeElement !== target) target.focus();
	});
}

// Vue Flow emits remove changes before the DOM is updated, so the focused element and its neighbors in tab order
// are captured here. Once removed, the focus moves to the closest preceding element that still exists, or the
// closest following one if there is no preceding element.
function restoreFocusAfterRemoval(): void {
	if (focusRestorePending || !flowChartContainer.value) return;

	const active = document.activeElement;
	const focusables = Array.from(
		flowChartContainer.value.querySelectorAll<HTMLElement | SVGElement>(
			'.vue-flow__node[tabindex], .vue-flow__edge[tabindex]'
		)
	);
	const index = focusables.findIndex((el) => el === active);
	if (index < 0) return;

	const candidates = [...focusables.slice(0, index).reverse(), ...focusables.slice(index + 1)];
	focusRestorePending = true;

	window.setTimeout(() => {
		focusRestorePending = false;
		if (active?.isConnected) return;
		candidates.find((el) => el.isConnected)?.focus();
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

function onNodeClick(e: NodeMouseEvent, rightClick?: boolean) {
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
		rightClick: rightClick,
	});
}

function onNodeContextMenu(e: NodeMouseEvent) {
	e.event.preventDefault();
	onNodeClick(e, true);
}

function onEdgeClick(e: EdgeMouseEvent, rightClick?: boolean) {
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
		rightClick: rightClick,
	});
}

function onEdgeKeyDown(event: KeyboardEvent, edgeId: string) {
	if (!flowChart.value || !['Enter', ' ', 'Escape'].includes(event.key)) return;

	const edge = flowChart.value.findEdge(edgeId);
	if (!edge) return;

	event.preventDefault();

	const selectable = edge.selectable ?? !!props.ui.elementsSelectable;

	if (event.key === 'Escape') {
		if (!selectable) return;
		flowChart.value.removeSelectedEdges([edge]);
		onSelectionChange();
		return;
	}

	if (selectable) {
		flowChart.value.addSelectedEdges([edge]);
	}

	const bounds = (event.currentTarget as Element).getBoundingClientRect();
	const viewPos = {
		x: bounds.left + bounds.width / 2,
		y: bounds.top + bounds.height / 2,
	};
	const panePos = getPanePosFromClientPos({ ...viewPos });

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

function onEdgeContextMenu(e: EdgeMouseEvent) {
	e.event.preventDefault();
	onEdgeClick(e, true);
}

function onPaneClick(e: MouseEvent, rightClick?: boolean) {
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
			rightClick: rightClick,
		});
	});
}

function onPaneContextMenu(e: MouseEvent) {
	onPaneClick(e, true);
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

const viewportAnimation = { duration: 200 };

function zoomIn() {
	flowChart.value?.zoomIn(viewportAnimation);
}

function zoomOut() {
	flowChart.value?.zoomOut(viewportAnimation);
}

function fitView() {
	flowChart.value?.fitView(viewportAnimation);
}

function onNagoEvent(invoke: NagoEvent): void {
	if (!(invoke instanceof CallRequested)) return;

	if (invoke.call instanceof FlowChartAutoLayout) autoLayout();
}

function init() {
	setNodesAndEdges();
	ConnectionHandler.addEventListener(onNagoEvent);
}

// Vue Flow only reflects the selection state as a CSS class, so it is mirrored to "aria-selected" for screen readers.
watchEffect(() => {
	const chart = flowChart.value;
	if (!chart) return;

	[...chart.getNodes, ...chart.getEdges].forEach((element) => {
		if (!element.domAttributes) return;
		const selectable = element.selectable ?? !!props.ui.elementsSelectable;
		const ariaSelected = selectable ? element.selected : undefined;
		if (element.domAttributes['aria-selected'] !== ariaSelected) {
			element.domAttributes['aria-selected'] = ariaSelected;
		}
	});
});

interface SelectedElement {
	kind: 'node' | 'edge';
	id: string;
	label: string;
}

const selectionAnnouncement = ref('');
let previousSelection = new Map<string, SelectedElement>();

function currentSelection(): Map<string, SelectedElement> {
	const selection = new Map<string, SelectedElement>();
	const chart = flowChart.value;
	if (!chart) return selection;

	chart.getSelectedNodes.forEach((n) =>
		selection.set(`node:${n.id}`, { kind: 'node', id: n.id, label: n.ariaLabel ?? n.id })
	);
	chart.getSelectedEdges.forEach((e) =>
		selection.set(`edge:${e.id}`, { kind: 'edge', id: e.id, label: e.ariaLabel ?? e.id })
	);
	return selection;
}

// Screen readers don't announce "aria-selected" changes on graphics roles, so changes are announced via a live region.
function announceSelectionChange(): void {
	const chart = flowChart.value;
	if (!chart) return;

	const selection = currentSelection();
	const messages: string[] = [];

	selection.forEach((element, key) => {
		if (!previousSelection.has(key)) messages.push(t('flowChart.selection.selected', { element: element.label }));
	});
	previousSelection.forEach((element, key) => {
		if (selection.has(key)) return;
		// removed elements are not announced as deselected
		const stillExists = element.kind === 'node' ? chart.findNode(element.id) : chart.findEdge(element.id);
		if (stillExists) messages.push(t('flowChart.selection.deselected', { element: element.label }));
	});

	previousSelection = selection;
	if (messages.length === 0) return;

	// resetting first ensures that an identical message is announced again
	selectionAnnouncement.value = '';
	nextTick(() => {
		selectionAnnouncement.value = messages.join(', ');
	});
}

watch(() => [...currentSelection().keys()].sort().join('|'), announceSelectionChange);

watch(() => props.ui.value, setNodesAndEdges, { deep: true });
watch(
	() => props.ui.movable,
	(movable) => {
		if (movable) {
			lockedViewport.value = undefined;
		} else if (flowChart.value) {
			lockedViewport.value = flowChart.value.getViewport();
		}
	},
	{ immediate: true }
);

onMounted(() => {
	if (!flowChartContainer.value) return;

	resizeObserver = new ResizeObserver(fitViewOnResize);
	resizeObserver.observe(flowChartContainer.value);
});

onUnmounted(() => {
	resizeObserver?.disconnect();
	ConnectionHandler.removeEventListener(onNagoEvent);
});

init();
</script>
<style scoped>
@reference '@/assets/css/tailwind.css';
.ui-flow-chart__canvas {
	@apply min-h-[inherit];
}

.ui-flow-chart {
	.aria-hidden {
		@apply absolute size-px opacity-0 pointer-events-none;
	}
}
</style>
