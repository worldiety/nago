<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { marginCSS } from '@/components/shared/padding';
import { GridCell } from '@/shared/protocol/ora/gridCell';

const props = defineProps<{
	ui: GridCell;
}>();

const style = computed<string>(() => {
	const styles: string[] = [];

	if (props.ui.rs) {
		styles.push(`grid-row-start: ${props.ui.rs}`);
	}

	if (props.ui.re) {
		styles.push(`grid-row-end: ${props.ui.re}`);
	}

	if (props.ui.cp) {
		styles.push(`grid-column: span ${props.ui.cp} / span ${props.ui.cp}`);
	}

	if (props.ui.rp) {
		styles.push(`grid-row: span ${props.ui.cp} / span ${props.ui.cp}`);
	}

	if (props.ui.cs) {
		styles.push(`grid-column-start: ${props.ui.cs}`);
	}

	if (props.ui.ce) {
		styles.push(`grid-column-end: ${props.ui.ce}`);
	}

	switch (props.ui.a) {
		case 'c':
			styles.push('place-self: center');
			break;
		case 'l':
			styles.push('place-self: center start');
			break;
		case 't':
			styles.push('place-self: center end');
			break;
		case 'u':
			styles.push('place-self: start center');
			break;
		case 'b':
			styles.push('place-self: end center');
			break;
		case 'ul':
			styles.push('place-self: start');
			break;
		case 'ut':
			styles.push('place-self: start end');
			break;
		case 'bl':
			styles.push('place-self: end start');
			break;
		case 'bt':
			styles.push('place-self: end end');
			break;
	}

	styles.push(...marginCSS(props.ui.p));

	return styles.join(';');
});
</script>

<template>
	<!-- gridcell -->
	<ui-generic v-if="props.ui.b" :ui="props.ui.b" :style="style" />
</template>
