<script lang="ts" setup>
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { Alignment } from '@/components/shared/alignments';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { fontCSS } from '@/components/shared/font';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue } from '@/components/shared/length';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { TextLayout } from '@/shared/protocol/ora/textLayout';
import { VStack } from '@/shared/protocol/ora/vStack';

const props = defineProps<{
	ui: TextLayout;
}>();

const serviceAdapter = useServiceAdapter();

function onClick() {
	if (props.ui.t) {
		serviceAdapter.executeFunctions(props.ui.t);
	}
}

// copy-paste me into UiText, UiVStack and UiHStack (or refactor me into some kind of generics-getter-setter-nightmare).
function commonStyles(): string[] {
	let styles = frameCSS(props.ui.f);

	styles.push(...borderCSS(props.ui.b));

	// other stuff
	styles.push(...paddingCSS(props.ui.p));
	styles.push(...fontCSS(props.ui.fn));

	return styles;
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles();

	return styles.join(';');
});

const clazz = computed<string>(() => {
	let classes = [];

	if (props.ui.t) {
		classes.push('cursor-pointer');
	}

	return classes.join(' ');
});
</script>

<template>
	<!-- textlayout -->
	<div v-if="!props.ui.iv" :class="clazz" :style="frameStyles" @click="onClick">
		<ui-generic v-for="ui in props.ui.c" :ui="ui" />
	</div>
</template>
