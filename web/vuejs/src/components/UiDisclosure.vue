<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import ArrowDownIcon from '@/assets/svg/chevron-down.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { bool2Str } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { CssClasses } from '@/shared/cssClasses';
import type { Disclosure } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Disclosure;
}>();

const serviceAdapter = useServiceAdapter();
const details = ref<HTMLDetailsElement>();

// The browser owns the open state of the details element. The template sets it only once, when the element is
// created: initialOpen is no reactive value, so Vue never patches the attribute again and a rendering does not reset
// what the user has chosen.
const initialOpen = !!props.ui.value;

// known is the open state, which the server has as far as we know: the value of the last rendering or the last report.
let known = initialOpen;

const classes = computed<string[]>(() => {
	const cls: string[] = [CssClasses.getOrCreate(frameCSS(props.ui.frame))];
	if (props.ui.separator) cls.push('separator');
	return cls;
});

function apply(open: boolean) {
	known = open;
	if (details.value && details.value.open !== open) {
		details.value.open = open; // the toggle event follows, but does not report it, because it is known
	}
}

// A rendering only applies the open state, if it changes compared to the previous rendering, e.g. the application
// opens another section for a deep link. A constant value never overrides what the user did.
watch(
	() => props.ui.value,
	(value) => apply(!!value)
);

// A stack patches its children by their position, so this instance may be reused for another disclosure, whose
// state does not belong to the previous one. Note the separate sources: a getter returning a new array would fire
// with every rendering and reset what the user has chosen.
watch([() => props.ui.inputValue, () => props.ui.id], () => apply(!!props.ui.value));

function onToggle() {
	if (!details.value) return;

	const open = details.value.open;
	if (open === known) return;

	known = open;
	if (props.ui.inputValue) {
		// in the background: the browser has already opened or closed the element
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), bool2Str(open)));
	}
}
</script>

<template>
	<details
		v-if="ui.summary"
		:id="ui.id || undefined"
		ref="details"
		class="disclosure"
		:class="classes"
		:open="initialOpen"
		@toggle="onToggle"
	>
		<summary class="summary">
			<span class="summary-content">
				<UiGeneric :ui="ui.summary" />
			</span>
			<span class="summary-icon" aria-hidden="true">
				<ArrowDownIcon />
			</span>
		</summary>
		<div v-if="ui.content" class="content">
			<UiGeneric :ui="ui.content" />
		</div>
	</details>
</template>

<style scoped>
@reference '@/assets/css/tailwind.css';
/*
 * The rules address the own parts of a disclosure with child combinators, because a disclosure within the content of
 * another one is a descendant as well, and must not take its open styles. The look follows the accordion.
 */
.disclosure {
	@apply text-M8;

	& > .summary {
		@apply w-full flex items-center py-3 cursor-pointer list-none;

		&::-webkit-details-marker {
			display: none;
		}

		&:focus-visible {
			@apply outline-solid outline-2 -outline-offset-2 outline-I0 rounded-lg;
		}

		& > .summary-content {
			@apply grow text-left;
		}

		& > .summary-icon {
			@apply px-4;

			svg {
				@apply size-3 duration-300;
			}
		}
	}

	& > .content {
		@apply w-full pb-2;
	}

	&.separator {
		@apply border-b border-M5;
	}

	&[open] > .summary > .summary-icon svg {
		@apply -scale-y-100;
	}
}
</style>
