<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { fontCSS } from '@/components/shared/font';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { FunctionCallRequested, TextLayout } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: TextLayout;
}>();

const serviceAdapter = useServiceAdapter();

function onClick() {
	if (props.ui.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	}
}

// copy-paste me into UiText, UiVStack and UiHStack (or refactor me into some kind of generics-getter-setter-nightmare).
function commonStyles(): string[] {
	let styles = frameCSS(props.ui.frame);

	styles.push(...borderCSS(props.ui.border));

	// other stuff
	styles.push(...paddingCSS(props.ui.padding));
	styles.push(...fontCSS(props.ui.font));

	return styles;
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles();

	return styles.join(';');
});

const clazz = computed<string>(() => {
	let classes = [];

	if (props.ui.action) {
		classes.push('cursor-pointer');
	}

	return classes.join(' ');
});
</script>

<template>
	<!-- textlayout -->
	<div v-if="!props.ui.invisible" :class="clazz" :style="frameStyles" @click="onClick">
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</div>
</template>
