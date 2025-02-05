<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import {ScrollView, ScrollViewAxisValues} from "@/shared/proto/nprotoc_gen";

const props = defineProps<{
	ui: ScrollView;
}>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	if (!props.ui.backgroundColor.isZero()) {
		styles.push(`background-color: ${props.ui.backgroundColor.value}`);
	}

	styles.push(...paddingCSS(props.ui.padding));

	return styles.join(';');
});

const classes = computed<string>(() => {
	const css: string[] = [];

	// note, that we defined its style in scrollbars.css
	switch (props.ui.axis.value) {
		case ScrollViewAxisValues.ScrollViewAxisHorizontal:
			css.push('overflow-x-auto', 'overflow-y-hidden');
			break;
		default:
			css.push('overflow-y-auto', 'overflow-x-hidden');
			break;
	}

	return css.join(' ');
});

const innerStyles = computed<string>(() => {
	let css = borderCSS(props.ui.border);

	switch (props.ui.axis.value) {
		case ScrollViewAxisValues.ScrollViewAxisHorizontal:
			css.push('width: max-content');
			break;
		default:
			css.push('height: max-content');
			break;
	}

	return css.join(';');
});

// note that we need the max-content hack, otherwise we get layout bugs at least for horizontal areas
</script>

<template v-if="props.ui.iv">
	<!-- UiScrollView -->
	<div :class="classes" :style="styles">
		<div :style="innerStyles">
			<UiGeneric v-if="ui.content" :ui="ui.content" />
		</div>
	</div>
</template>
