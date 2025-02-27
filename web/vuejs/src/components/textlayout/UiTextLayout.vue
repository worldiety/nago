<script lang="ts" setup>
import {computed} from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import {borderCSS} from '@/components/shared/border';
import {fontCSS} from '@/components/shared/font';
import {frameCSS} from '@/components/shared/frame';
import {paddingCSS} from '@/components/shared/padding';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {TextLayout} from '@/shared/protocol/ora/textLayout';
import {FunctionCallRequested} from "@/shared/proto/nprotoc_gen";
import {nextRID} from "@/eventhandling";

const props = defineProps<{
	ui: TextLayout;
}>();

const serviceAdapter = useServiceAdapter();

function onClick() {
	if (props.ui.t) {
		serviceAdapter.sendEvent(new FunctionCallRequested(
			props.ui.t,
			nextRID(),
		));
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
		<ui-generic v-for="ui in props.ui.c" :ui="ui"/>
	</div>
</template>
