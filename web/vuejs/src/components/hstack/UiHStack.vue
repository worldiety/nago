<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed, ref} from 'vue';
import {HStack} from "@/shared/protocol/ora/hStack";
import {frameCSS} from "@/components/shared/frame";
import {Alignment} from "@/components/shared/alignments";
import {cssLengthValue} from "@/components/shared/length";
import {paddingCSS} from "@/components/shared/padding";
import {colorValue} from "@/components/shared/colors";
import {fontCSS} from "@/components/shared/font";
import {borderCSS} from "@/components/shared/border";
import {useServiceAdapter} from "@/composables/serviceAdapter";

const props = defineProps<{
	ui: HStack;
}>();

const hover = ref(false);
const pressed = ref(false);
const focused = ref(false);
const focusable = ref(false);
const focusVisible = ref(false);
const serviceAdapter = useServiceAdapter();

function onClick(event: Event) {
	if (props.ui.t) {
		event.stopPropagation()
		serviceAdapter.executeFunctions(props.ui.t);
	}
}

function onKeydown(event: KeyboardEvent) {
	if (props.ui.t) {
		event.stopPropagation()
		if (event.code === "Enter" || event.code === 'Space') {
			serviceAdapter.executeFunctions(props.ui.t);
		}

	}
}


function checkFocusVisible(event: Event) {
	const element = event.target as HTMLElement;
	focusVisible.value = element.matches(":focus-visible");
}

// copy-paste me into UiText, UiVStack and UiHStack (or refactor me into some kind of generics-getter-setter-nightmare).
function commonStyles(): string[] {
	let styles = frameCSS(props.ui.f)

	// background handling
	if (props.ui.pgc && pressed.value) {
		styles.push(`background-color: ${colorValue(props.ui.pgc)}`)
	} else {
		if (props.ui.hgc) {
			if (hover.value) {
				styles.push(`background-color: ${colorValue(props.ui.hgc)}`)
			} else {
				styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
			}
		} else {
			styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
		}
	}

	if (props.ui.t) {
		focusable.value = true
	}

	if (props.ui.fbc) {
		focusable.value = true;
		if (focused.value && !pressed.value) {
			styles.push(`background-color: ${colorValue(props.ui.fbc)}`)
		}
	}

	// border handling
	if (props.ui.pb && pressed.value) {
		styles.push(...borderCSS(props.ui.pb))
	} else {
		if (props.ui.hb) {
			if (hover.value) {
				styles.push(...borderCSS(props.ui.hb))
			} else {
				styles.push(...borderCSS(props.ui.b))
			}
		} else {
			styles.push(...borderCSS(props.ui.b))
		}
	}

	if (props.ui.fb) {
		focusable.value = true;
		if (focusVisible.value) {
			styles.push(...borderCSS(props.ui.fb))
		}
	}


	// other stuff
	styles.push(...paddingCSS(props.ui.p))
	styles.push(...fontCSS(props.ui.fn))

	if (focusVisible.value) {
		styles.push("outline: 2px solid black") // always apply solid and never auto. Auto will create random broken effects on firefox and chrome
	}

	return styles
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles()


	if (props.ui.g) {
		styles.push(`column-gap:${cssLengthValue(props.ui.g)}`)
	}


	return styles.join(";")
});

const StyleButtonPrimary = "p"
const StyleButtonSecondary = "s"
const StyleButtonTertiary = "t"


const clazz = computed<string>(() => {

	let classes = ["inline-flex"];
	switch (props.ui.a) {
		case Alignment.Leading:
			classes.push("justify-start", "items-center")
			break
		case Alignment.Trailing:
			classes.push("justify-end", "items-center")
			break
		case Alignment.Center:
			classes.push("justify-center", "items-center")
			break
		case Alignment.TopLeading:
			classes.push("justify-start", "items-start")
			break
		case Alignment.BottomLeading:
			classes.push("justify-start", "items-end")
			break
		case Alignment.TopTrailing:
			classes.push("justify-end", "items-start")
			break
		case Alignment.Top:
			classes.push("justify-center", "items-start")
			break
		case Alignment.BottomTrailing:
			classes.push("justify-end", "items-end")
			break
		case Alignment.Bottom:
			classes.push("justify-center", "items-end")
			break
		default:
			classes.push("justify-center", "items-center")
			break

	}

	if (props.ui.t) {
		classes.push("cursor-pointer")
	}

	if (props.ui.w) {
		classes.push("flex-wrap")
	}

	switch (props.ui.s) {
		case StyleButtonPrimary:
			classes.push("button-primary")
			break
		case StyleButtonSecondary:
			classes.push("button-secondary")
			break
		case StyleButtonTertiary:
			classes.push("button-tertiary")
			break
	}


	// preset special round icon mode in buttons
	if (props.ui.s) {
		if (props.ui.c && props.ui.c.length == 1 && props.ui.c[0].type == "I") {
			classes.push('!p-0', '!w-10')
		}
	}

	return classes.join(" ")
});

</script>

<template>
	<!-- hstack -->
	<div v-if="!props.ui.s && !props.ui.iv" :class="clazz" :style="frameStyles" @mouseover="hover = true"
			 @mouseleave="hover = false"
			 @mousedown="pressed = true" @mouseup="pressed = false" @mouseout="pressed = false" @focusin="focused = true"
			 :title="props.ui.al"
			 @focusout="focused = false;focusVisible=false" :tabindex="focusable?0:-1" @click="onClick" @keydown="onKeydown"
			 @focus="checkFocusVisible">
		<ui-generic v-for="ui in props.ui.c" :ui="ui"/>
	</div>

	<button :disabled="props.ui.d" v-if="props.ui.s && !props.ui.iv" :class="clazz" :style="frameStyles" @click="onClick"
					:title="props.ui.al">
		<ui-generic v-for="ui in props.ui.c" :ui="ui"/>
	</button>
</template>
