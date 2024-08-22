<script lang="ts" setup>
import {computed} from 'vue';
import {frameCSS} from "@/components/shared/frame";
import {Spacer} from "@/shared/protocol/ora/spacer";
import {borderCSS} from "@/components/shared/border";

const props = defineProps<{
	ui: Spacer;
}>();


const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.b)
	styles.push(...frameCSS(props.ui.f))
	if (props.ui.bgc) {
		styles.push(`background-color: ${props.ui.bgc}`)
	}

	styles.push("object-fit: cover")

	return styles.join(";")
});
</script>

<template v-if="props.ui.children">
	<!-- spacer -->
	<div class="grow shrink" :style="styles">
	</div>
</template>
