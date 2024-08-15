<script lang="ts" setup>
import {computed, ref} from 'vue';
import type {Text} from "@/shared/protocol/ora/text";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {isNil} from "@/shared/protocol/util";
import {frameCSS} from "@/components/shared/frame";
import {paddingCSS} from "@/components/shared/padding";
import {cssLengthValue} from "@/components/shared/length";
import {colorValue} from "@/components/shared/colors";
import {fontCSS} from "@/components/shared/font";
import {borderCSS} from "@/components/shared/border";

const props = defineProps<{
	ui: Text;
}>();

const hover = ref(false);
const pressed = ref(false);
const focused = ref(false);
const focusable = ref(false);
const serviceAdapter = useServiceAdapter();

function onClick() {
	if (props.ui.t){
		serviceAdapter.executeFunctions(props.ui.t);
	}
}


const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.f)
	if (props.ui.c) {
		styles.push(`color: ${colorValue(props.ui.c)}`)
	}

	if (props.ui.bgc) {
		styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
	}

	styles.push(...borderCSS(props.ui.b))
	styles.push(...paddingCSS(props.ui.p))
	styles.push(...fontCSS(props.ui.o))

	switch (props.ui.a){
		case "s":
			styles.push("text-align: start")
			break
		case "e":
			styles.push("text-align: end")
			break
		case "c":
			styles.push("text-align: center")
			break
		case "j":
			styles.push("text-align: justify","text-justify: inter-character") // inter-character just looks so much better
			break
	}

	if (props.ui.t){
		styles.push("cursor: pointer")
	}

	return styles.join(";")
});


</script>

<template>
	<span v-if="!ui.i" :style="styles" @click="onClick" >{{
			props.ui.v
		}}</span>
</template>
