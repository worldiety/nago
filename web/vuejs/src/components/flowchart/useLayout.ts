import dagre from '@dagrejs/dagre';
import { Node, Position, useVueFlow, VueFlow } from '@vue-flow/core';
import { ref } from 'vue';
import { OrientationValues } from '@/shared/proto/nprotoc_gen';

/**
 * Composable to run the layout algorithm on the graph.
 * It uses the `dagre` library to calculate the layout of the nodes and edges.
 */
export function useLayout() {
	const { findNode } = useVueFlow();

	const graph = ref(new dagre.graphlib.Graph());

	const previousDirection = ref(OrientationValues.Horizontal);

	function layout(flowChart: InstanceType<typeof VueFlow>, direction: OrientationValues): Node[] {
		// we create a new graph instance, in case some nodes/edges were removed, otherwise dagre would act as if they were still there
		const dagreGraph = new dagre.graphlib.Graph();
		const nodes = flowChart.nodes;
		const edges = flowChart.edges;

		graph.value = dagreGraph;

		dagreGraph.setDefaultEdgeLabel(() => ({}));

		const isHorizontal = direction === OrientationValues.Horizontal;
		dagreGraph.setGraph({ rankdir: direction === OrientationValues.Horizontal ? 'LR' : 'TB' });

		previousDirection.value = direction;

		for (const node of nodes) {
			// if you need width+height of nodes for your layout, you can use the dimensions property of the internal node (`GraphNode` type)
			const graphNode = findNode(node.id);
			if (!graphNode) continue;

			dagreGraph.setNode(node.id, {
				width: graphNode.dimensions.width || 150,
				height: graphNode.dimensions.height || 50,
			});
		}

		for (const edge of edges) {
			dagreGraph.setEdge(edge.source, edge.target);
		}

		dagre.layout(dagreGraph);

		// set nodes with updated positions
		return nodes.map((node) => {
			const nodeWithPosition = dagreGraph.node(node.id);

			return {
				...node,
				data: { ...node.data, position: { x: nodeWithPosition.x, y: nodeWithPosition.y } },
				targetPosition: isHorizontal ? Position.Left : Position.Top,
				sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
				position: { x: nodeWithPosition.x, y: nodeWithPosition.y },
			};
		});
	}

	return { graph, layout, previousDirection };
}
