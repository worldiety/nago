<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { marginCSS } from '@/components/shared/padding';
import { AlignmentValues, GridCell } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: GridCell;
}>();

const style = computed<string>(() => {
	const styles: string[] = [];

	if (!props.ui.rowStart.isZero()) {
		styles.push(`grid-row-start: ${props.ui.rowStart.value}`);
	}

	if (!props.ui.rowEnd.isZero()) {
		styles.push(`grid-row-end: ${props.ui.rowEnd.value}`);
	}

	if (!props.ui.colSpan.isZero()) {
		styles.push(`grid-column: span ${props.ui.colSpan.value} / span ${props.ui.colSpan.value}`);
	}

	if (!props.ui.rowSpan.isZero()) {
		styles.push(`grid-row: span ${props.ui.rowSpan.value} / span ${props.ui.rowSpan.value}`);
	}

	if (!props.ui.colStart.isZero()) {
		styles.push(`grid-column-start: ${props.ui.colStart.value}`);
	}

	if (!props.ui.colEnd.isZero()) {
		styles.push(`grid-column-end: ${props.ui.colEnd.value}`);
	}

	switch (props.ui.alignment.value) {
		case AlignmentValues.Center:
			// TODO we have inherited a strange behavior here: in the ora protocol version we were undefined and did not apply the default behavior which breaks the grid area
			//styles.push('place-self: center');
			break;
		case AlignmentValues.Leading:
			styles.push('place-self: center start');
			break;
		case AlignmentValues.Trailing:
			styles.push('place-self: center end');
			break;
		case AlignmentValues.Top:
			styles.push('place-self: start center');
			break;
		case AlignmentValues.Bottom:
			styles.push('place-self: end center');
			break;
		case AlignmentValues.TopLeading:
			styles.push('place-self: start');
			break;
		case AlignmentValues.TopTrailing:
			styles.push('place-self: start end');
			break;
		case AlignmentValues.BottomLeading:
			styles.push('place-self: end start');
			break;
		case AlignmentValues.BottomTrailing:
			styles.push('place-self: end end');
			break;
	}

	styles.push(...marginCSS(props.ui.padding));

	return styles.join(';');
});
</script>

<template>
	<!-- gridcell -->
	<ui-generic v-if="props.ui.body" :ui="props.ui.body" :style="style" />
</template>
