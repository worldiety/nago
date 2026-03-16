<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import ArrowDownIcon from '@/assets/svg/arrowDown.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { bool2Str, randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
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

const classes = computed<string[]>(() => {
	const styles = frameCSS(props.ui.frame);
	const cls: string[] = [CssClasses.getOrCreate(styles)];
	if (props.ui.value) cls.push('open');
	return cls;
});

function calcBodyHeight() {
	if (!props.ui.value || !bodyDummy.value) {
		bodyHeight.value = 0;
		return;
	}

	bodyHeight.value = bodyDummy.value.getBoundingClientRect().height;
}

function toggle(value?: boolean) {
	serviceAdapter.sendEvent(
		new UpdateStateValueRequested(
			props.ui.inputValue,
			0,
			nextRID(),
			bool2Str(value !== undefined ? value : !props.ui.value)
		)
	);
}

onMounted(() => {
	setTimeout(calcBodyHeight, 20); // TODO: Find better way to render correct initial state
	addEventListener('resize', calcBodyHeight);
	watch(() => props.ui.value, calcBodyHeight);
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
			<div class="body" :style="`height: ${bodyHeight}px;`" :inert="!ui.value">
				<UiGeneric :ui="ui.content" />
			</div>
			<div ref="bodyDummy" class="body-dummy" inert aria-hidden="true">
				<UiGeneric :ui="ui.content" />
			</div>
		</div>
	</div>
</template>

<style scoped>
.accordion-container {
	.accordion {
		@apply relative w-full;

		.header {
			@apply w-full flex items-center py-1;

			.header-content {
				@apply grow text-left;
			}

			.header-icon {
				@apply px-4;

				svg {
					@apply size-4 duration-100;
				}
			}
		}

		.body {
			@apply w-full overflow-hidden duration-100;
		}

		.body-dummy {
			@apply absolute left-0 w-full opacity-0 pointer-events-none;
		}
	}

	&.open {
		.accordion {
			.header {
				.header-icon {
					svg {
						@apply -scale-y-100;
					}
				}
			}
		}
	}
}
</style>
