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

	styles.push("object-fit: cover")

	return styles.join(";")
})
</script>

<template>
	<img v-if="!ui.iv" class="h-auto max-w-full" :src="props.ui.u" :alt="props.ui.al" :style="styles"/>
</template>
