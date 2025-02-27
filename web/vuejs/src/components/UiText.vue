<script lang="ts" setup>
import { computed, ref } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { fontCSS } from '@/components/shared/font';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { FunctionCallRequested, TextAlignmentValues, TextView } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: TextView;
}>();

const hover = ref(false);
const pressed = ref(false);
const focused = ref(false);
const focusable = ref(false);
const serviceAdapter = useServiceAdapter();

function onClick() {
	if (props.ui.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	}
}

const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);
	if (props.ui.color) {
		styles.push(`color: ${colorValue(props.ui.color)}`);
	}

	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	}

	styles.push(...borderCSS(props.ui.border));
	styles.push(...paddingCSS(props.ui.padding));
	styles.push(...fontCSS(props.ui.font));
	styles.push('white-space:pre-wrap'); // TODO not sure if this is the intentional effect for all platforms

	switch (props.ui.textAlignment) {
		case TextAlignmentValues.TextAlignStart:
			styles.push('text-align: start');
			break;
		case TextAlignmentValues.TextAlignEnd:
			styles.push('text-align: end');
			break;
		case TextAlignmentValues.TextAlignCenter:
			styles.push('text-align: center');
			break;
		case TextAlignmentValues.TextAlignJustify:
			styles.push('text-align: justify', 'text-justify: inter-character'); // inter-character just looks so much better
			break;
	}


	if (props.ui.action) {
		styles.push('cursor: pointer');
	}

	return styles.join(';');
});
</script>

<template>
	<span v-if="!ui.invisible" :style="styles" @click="onClick"
		>{{ props.ui.value }} <br v-if="!ui.invisible && ui.lineBreak" />
	</span>
</template>
