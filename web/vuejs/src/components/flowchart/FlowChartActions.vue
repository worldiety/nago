<template>
	<Panel
		v-if="toolbar.actions"
		class="flow-chart-actions"
		:class="{ vertical: toolbar.orientation === OrientationValues.Vertical }"
		:position="position"
	>
		<button
			v-if="toolbar.actions.value.includes(FlowChartToolbarActionValues.FlowChartToolbarActionAutoLayout)"
			:title="$t('flowChart.actions.autoLayout')"
			@click="onAutoLayout"
		>
			<IconArrowLeftRight v-if="orientation === OrientationValues.Horizontal" />
			<IconArrowUpDown v-else />
		</button>
	</Panel>
</template>
<script lang="ts" setup>
import {
	FlowChartToolbar,
	FlowChartToolbarActionValues,
	FlowChartToolbarPositionValues,
	OrientationValues,
} from '@/shared/proto/nprotoc_gen';
import { Panel, PanelPositionType, VueFlow } from '@vue-flow/core';
import IconArrowLeftRight from '@/assets/svg/arrow-left-right.svg';
import IconArrowUpDown from '@/assets/svg/arrow-up-down.svg';
import { computed } from 'vue';

interface Props {
	toolbar: FlowChartToolbar;
	orientation: OrientationValues;
	flowChart: InstanceType<typeof VueFlow>;
}

interface Emits {
	(e: 'action:autoLayout'): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const position = computed<PanelPositionType>(() => {
	switch (props.toolbar.position) {
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionCenterRight:
			return 'center-right' as PanelPositionType;
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionBottomRight:
			return 'bottom-right';
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionBottomCenter:
			return 'bottom-center';
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionBottomLeft:
			return 'bottom-left';
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionCenterLeft:
			return 'center-left' as PanelPositionType;
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionTopLeft:
			return 'top-left';
		case FlowChartToolbarPositionValues.FlowChartToolbarPositionTopCenter:
			return 'top-center';
		default:
			return 'top-right';
	}
});

function onAutoLayout() {
	emit('action:autoLayout');
}
</script>
<style scoped>
.flow-chart-actions {
	@apply flex gap-2 m-0 p-4;

	button {
		@apply rounded-lg p-2 border border-M3 bg-M2 hover:bg-M3;

		svg {
			@apply size-5 fill-current;
		}
	}

	&.vertical {
		@apply flex-col;
	}

	&.top.right {
		@apply top-0 right-0 bottom-auto left-auto;
	}

	&.center.right {
		@apply top-1/2 right-0 bottom-auto left-auto -translate-y-1/2;
	}

	&.bottom.right {
		@apply top-auto right-0 bottom-0 left-auto;
	}

	&.bottom.center {
		@apply top-auto right-auto bottom-0 left-1/2 -translate-x-1/2;
	}

	&.bottom.left {
		@apply top-auto right-auto bottom-0 left-0;
	}

	&.center.left {
		@apply top-1/2 right-auto bottom-auto left-0 -translate-y-1/2;
	}

	&.top.left {
		@apply top-0 right-auto bottom-auto left-0;
	}

	&.top.center {
		@apply top-0 right-auto bottom-auto left-1/2 -translate-x-1/2;
	}
}
</style>
