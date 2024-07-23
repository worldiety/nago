<script lang="ts" setup>
import type {Image} from "@/shared/protocol/ora/image";
import {computed} from "vue";
import {borderCSS} from "@/components/shared/border";
import {frameCSS} from "@/components/shared/frame";
import {paddingCSS} from "@/components/shared/padding";

const props = defineProps<{
	ui: Image;
}>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.b)
	styles.push(...frameCSS(props.ui.f))
	styles.push(...paddingCSS(props.ui.p))

	if (!props.ui.s) {
		styles.push("object-fit: cover")
	}

	return styles.join(";")
})

const rewriteSVG = computed<string>(() => {
	if (!props.ui.s){
		return ""
	}

		return props.ui.s.replace('<svg ', '<svg style="width: 100px; height: 100px; color: red;" ');
})
</script>

<template>
	<img v-if="!ui.iv && !ui.s" class="h-auto max-w-full" :src="props.ui.u" :alt="props.ui.al" :style="styles"/>
	<div v-if="props.ui.s" :style="styles" class="" v-html="rewriteSVG"></div>
</template>
