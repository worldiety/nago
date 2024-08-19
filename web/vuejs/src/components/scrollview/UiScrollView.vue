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
		case "v":
			css.push("overflow-x-auto")
		case "h":
			css.push("overflow-y-auto")
	}

	return css.join(" ")
})
</script>


<template v-if="props.ui.iv">
	<div :class="classes" :style="styles">
		<UiGeneric v-if="ui.c" :ui="ui.c"/>

	</div>
</template>
