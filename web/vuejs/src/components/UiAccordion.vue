<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import ArrowDownIcon from '@/assets/svg/chevron-down.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { OptimisticToggle } from '@/components/shared/optimisticToggle';
import { bool2Str, randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID, renderedRID } from '@/eventhandling';
import { CssClasses } from '@/shared/cssClasses';
import type { Accordion } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Accordion;
}>();

const id = randomStr(16);
const serviceAdapter = useServiceAdapter();

const bodyHeight = ref(0);
const bodyDummy = ref<HTMLDivElement>();

// The accordion opens and closes in the browser at once and sends the new state afterwards. Until a rendering shows,
// that the server has processed it, renderings keep the local state, see OptimisticToggle.
const sync = new OptimisticToggle(!!props.ui.value, props.ui.inputValue);
const open = ref<boolean>(sync.value);

// every rendering decides, also one with an unchanged value: the server may not have taken the toggle
watch(
	() => props.ui,
	(ui) => {
		open.value = sync.adopt(!!ui.value, ui.inputValue, renderedRID());
	}
);

const classes = computed<string[]>(() => {
	const styles = frameCSS(props.ui.frame);
	const cls: string[] = [CssClasses.getOrCreate(styles)];
	if (open.value) cls.push('open');
	if (props.ui.small) cls.push('small');
	if (props.ui.separator) cls.push('separator');
	return cls;
});

function calcBodyHeight() {
	if (!open.value || !bodyDummy.value) {
		bodyHeight.value = 0;
		return;
	}

	bodyHeight.value = bodyDummy.value.getBoundingClientRect().height;
}

function toggle(value?: boolean) {
	const next = value !== undefined ? value : !open.value;
	if (next === open.value) return;

	const rid = nextRID();
	sync.toggle(next, rid);
	open.value = next;
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, rid, bool2Str(next)));
}

function observeBodyHeightChange() {
	if (!bodyDummy.value) return;
	const observer = new ResizeObserver(calcBodyHeight);
	observer.observe(bodyDummy.value);
}

onMounted(() => {
	setTimeout(calcBodyHeight, 20); // TODO: Find better way to render correct initial state
	addEventListener('resize', calcBodyHeight);
	watch(open, calcBodyHeight);
	observeBodyHeightChange();
});

onUnmounted(() => {
	removeEventListener('resize', calcBodyHeight);
});
</script>

<template>
	<div v-if="ui.header && ui.content" :id="id" class="accordion-container" :class="classes">
		<div class="accordion">
			<button
				class="header"
				:aria-controls="`body-${id}`"
				:aria-expanded="open"
				@click="toggle()"
				@keydown.down.exact="toggle(true)"
				@keydown.up.exact="toggle(false)"
			>
				<span class="header-content">
					<UiGeneric :ui="ui.header" />
				</span>
				<span class="header-icon">
					<ArrowDownIcon />
				</span>
			</button>
			<div :id="`body-${id}`" class="body" :style="`height: ${bodyHeight}px;`" :inert="!open">
				<div class="body-inner">
					<UiGeneric :ui="ui.content" />
				</div>
			</div>
			<div ref="bodyDummy" class="body-dummy" inert aria-hidden="true">
				<UiGeneric :ui="ui.content" />
			</div>
		</div>
	</div>
</template>

<style scoped>
@reference '@/assets/css/tailwind.css';
/*
 * Every rule addresses the own parts of an accordion with child combinators: an accordion within the body of another
 * one is a descendant as well, and must neither open with its parent nor take its size.
 */
.accordion-container {
	@apply text-M8;

	& > .accordion {
		@apply relative w-full;

		& > .header {
			@apply w-full flex items-center py-3;

			& > .header-content {
				@apply grow text-left;
			}

			& > .header-icon {
				@apply px-4;

				svg {
					@apply size-3 duration-300;
				}
			}
		}

		& > .body {
			@apply w-full overflow-clip duration-300 opacity-0 pt-2 px-2 -ml-2 -mt-2 pointer-events-none;

			& > .body-inner {
				@apply w-full pointer-events-auto;
			}
		}

		& > .body-dummy {
			@apply absolute left-0 w-full opacity-0 pointer-events-none p-2;
		}
	}

	&.small > .accordion > .header {
		@apply py-0;

		& > .header-icon {
			@apply px-2;

			svg {
				@apply size-2 duration-300;
			}
		}
	}

	&.separator > .accordion {
		@apply border-b border-M5;
	}

	&.open > .accordion {
		& > .header > .header-icon svg {
			@apply -scale-y-100;
		}

		& > .body {
			@apply pb-2 opacity-100;
		}
	}
}
</style>
