<script lang="ts" setup>
import {computed} from 'vue';
import {frameCSS} from "@/components/shared/frame";
import {borderCSS} from "@/components/shared/border";
import {ScrollView} from "@/shared/protocol/ora/scrollView";
import UiGeneric from "@/components/UiGeneric.vue";
import {paddingCSS} from "@/components/shared/padding";

const props = defineProps<{
	ui: ScrollView;
}>();


const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.b)
	styles.push(...frameCSS(props.ui.f))
	if (props.ui.bgc) {
		styles.push(`background-color: ${props.ui.bgc}`)
	}

	styles.push(...paddingCSS(props.ui.p))


	return styles.join(";")
});

const classes = computed<string>(() => {
	const css: string[] = [];

	// note, that we defined its style in scrollbars.css
	switch (props.ui.a) {

		case "h":
			css.push("overflow-x-auto", "overflow-y-hidden")
			break
		default:
			css.push("overflow-y-auto", "overflow-x-hidden")
			break
	}

	return css.join(" ")
})

const innerStyles = computed<string>(() => {
	let css = borderCSS(props.ui.b)

	switch (props.ui.a) {

		case "h":
			css.push("width: max-content")
			break
		default:
			css.push("height: max-content")
			break
	}

	return css.join(";")
});

// note that we need the max-content hack, otherwise we get layout bugs at least for horizontal areas
</script>


<template v-if="props.ui.iv">
	<!-- UiScrollView -->
	<div :class="classes" :style="styles">
		<div :style="innerStyles">
			<UiGeneric v-if="ui.c" :ui="ui.c"/>
		</div>

	</div>
</template>
