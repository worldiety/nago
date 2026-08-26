<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div :id="id" ref="alertNotifications" class="alert-notifications" :class="{ stacked: stacked }">
		<div v-if="notifications.length" class="actions">
			<button class="button-tertiary" @click.stop="onClickCloseAll">
				<CloseIcon />
				{{ $t('alertNotifications.actions.closeAll') }}
			</button>
			<button v-if="notifications.length >= 3" class="button-tertiary" @click.stop="onClickToggleExpand">
				<ArrowDownIcon :class="{ '-scale-y-100': !stacked }" />
				<span v-if="expanded">
					{{ $t('alertNotifications.actions.collapse') }}
				</span>
				<span v-else>
					{{ $t('alertNotifications.actions.expand') }}
				</span>
			</button>
		</div>
		<div
			ref="notificationContainer"
			class="notifications overflow-y-auto"
			:class="containerOverflowClass"
			@scroll="calcContainerOverflowClass"
		>
			<div class="notifications-inner" :style="innerStyles">
				<div
					v-for="(noti, i) in notifications"
					ref="notificationElements"
					:key="getNotificationKey(noti)"
					class="notification"
					:style="getNotificationStyles(i)"
					:inert="stacked && i > 0"
				>
					<UiGeneric :ui="noti" />
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { randomStr } from '@/components/shared/util';
import { AlertNotifications, Component, FunctionCallRequested, Stack } from '@/shared/proto/nprotoc_gen';
import UiGeneric from '@/components/UiGeneric.vue';
import { nextRID } from '@/eventhandling';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { computed, onMounted, ref, watch } from 'vue';
import ArrowDownIcon from '@/assets/svg/chevron-down.svg';
import CloseIcon from '@/assets/svg/close.svg';

const props = defineProps<{
	ui: AlertNotifications;
}>();

const GAP = 16;
const SCALE_DIFF = 0.1;
const OPACITY_DIFF = 0.2;

const id = randomStr(16);
const serviceAdapter = useServiceAdapter();

const resizeUpdateKey = ref(0);
const resizeTimeout = ref();

const alertNotifications = ref<HTMLDivElement>();
const expanded = ref(false);
const notificationContainer = ref<HTMLDivElement>();
const notificationElements = ref<HTMLElement[]>();
const stacked = ref(false);
const containerOverflowClass = ref('');
const resizing = ref(false);

const transitionDuration = computed<number>(() => {
	if (resizing.value) return 0;
	return 300;
});

const notifications = computed<Stack[]>(() => {
	const filtered = (props.ui.notifications?.value.filter((n) => n instanceof Stack) as Stack[]) ?? [];
	return filtered.slice().reverse();
});

const innerStyles = computed<string | undefined>(() => {
	if (stacked.value || !notificationTops.value.length) return undefined;

	let top = 0;
	if (stacked.value) {
		top = notificationTops.value[0];
	} else {
		notificationTops.value.forEach((t) => {
			if (t > top) top = t;
		});
	}

	return `height: ${top}px;`;
});

const notificationTops = computed<number[]>(() => {
	dummyFunc(resizeUpdateKey.value); // This is needed to force recomputing on window resize (resize key change)

	const tops: number[] = [];
	notificationElements.value
		?.slice()
		.reverse()
		.forEach((noti, i) => {
			if (!notificationElements.value) return;

			if (i === 0) {
				tops.push(noti.clientHeight);
			} else {
				if (stacked.value) {
					if (i < 3) tops.push(tops[i - 1] + GAP);
					else tops.push(tops[0]);
				} else {
					tops.push(tops[i - 1] + GAP + noti.clientHeight);
				}
			}
		});
	return tops;
});

const notificationScales = computed<number[]>(() => {
	const tops: number[] = [];
	notificationElements.value?.forEach((_noti, i) => {
		if (!notificationElements.value) return;

		if (i === 0 || !stacked.value) tops.push(1);
		else tops.push(Math.max(0, 1 - i * SCALE_DIFF));
	});
	return tops;
});

const notificationOpacities = computed<number[]>(() => {
	const tops: number[] = [];
	notificationElements.value?.forEach((_noti, i) => {
		if (!notificationElements.value) return;

		if (i === 0 || !stacked.value) {
			tops.push(1);
		} else {
			if (i < 3) tops.push(1 - i * OPACITY_DIFF);
			else tops.push(0);
		}
	});
	return tops;
});

function dummyFunc(key: number) {
	return key;
}

function getNotificationKey(noti: Component): string | undefined {
	if (noti instanceof Stack) {
		return noti.id;
	}

	return undefined;
}

function getNotificationStyles(i: number): string {
	const styles: string[] = [
		`--notification-opacity: ${notificationOpacities.value[i]}`,
		`top: ${notificationTops.value[i]}px`,
		`transform: translate(-50%, -100%) scale(${notificationScales.value[i]})`,
		`opacity: var(--notification-opacity)`,
		`z-index: ${(notificationElements.value?.length ?? 0) - i}`,
	];

	if (!resizing.value) styles.push(`transition-duration: ${transitionDuration.value}ms`);

	return styles.join('; ');
}

function onClickToggleExpand() {
	expanded.value = !expanded.value;
	setStacked();
}

function onClickCloseAll() {
	serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.closeAll, nextRID()));
}

function setStacked() {
	if (notificationElements.value) {
		stacked.value = !expanded.value && (notificationElements.value?.length ?? 0) >= 3;
	}
}

function onNotificationsChanged() {
	setTimeout(() => {
		setStacked();

		if (alertNotifications.value) {
			const notiElements = alertNotifications.value.querySelectorAll('.notification');
			notiElements.forEach((noti) => {
				noti.classList.add('animated');
			});
		}
	}, 50);
}

function calcContainerOverflowClass() {
	if (!notificationContainer.value) return;

	if (stacked.value) {
		containerOverflowClass.value = '';
		return;
	}

	const overflowTop = notificationContainer.value.scrollTop > 0;
	const overflowBottom =
		notificationContainer.value.scrollHeight - notificationContainer.value.scrollTop >
		notificationContainer.value.clientHeight;
	if (overflowTop && overflowBottom) {
		containerOverflowClass.value = 'overflow-both';
		return;
	}
	if (overflowTop) {
		containerOverflowClass.value = 'overflow-top';
		return;
	}
	if (overflowBottom) {
		containerOverflowClass.value = 'overflow-bottom';
		return;
	}
}

function onWindowResize() {
	resizeUpdateKey.value++;
	if (resizeTimeout.value) clearTimeout(resizeTimeout.value);
	resizeTimeout.value = setTimeout(() => (resizing.value = false), 100);
	resizing.value = true;
}

function observeInnerContainerSize() {
	const containerInner = notificationContainer.value?.querySelector('.notifications-inner');
	if (!containerInner) return;

	const observer = new ResizeObserver(calcContainerOverflowClass);
	observer.observe(containerInner);
}

watch(() => props.ui.notifications, onNotificationsChanged, { deep: true });
addEventListener('resize', onWindowResize);

onMounted(() => {
	observeInnerContainerSize();
});
</script>
<style scoped>
.alert-notifications {
	@apply flex flex-col gap-4 h-full pointer-events-none;

	.actions {
		@apply flex items-center justify-start flex-row-reverse gap-4;

		& > button {
			@apply pointer-events-auto;

			& > svg {
				@apply size-3 duration-100;
			}
		}
	}

	.notifications {
		@apply overflow-y-auto max-h-full grow -mr-3 pl-6;
		scrollbar-gutter: stable;

		.notifications-inner {
			@apply relative flex flex-col gap-4 pointer-events-auto;

			.notification {
				@apply absolute pointer-events-auto top-0 left-1/2 -translate-x-1/2 origin-bottom max-w-full;
				animation: notification-create 200ms;

				&:first-child {
					@apply relative;
				}

				&:not(.animated) {
					@apply !duration-0;
				}
			}
		}

		&.overflow-top {
			mask-image: linear-gradient(transparent 0, black 1.5rem);
		}

		&.overflow-bottom {
			mask-image: linear-gradient(to top, transparent 0, black 1.5rem);
		}

		&.overflow-both {
			mask-image: linear-gradient(transparent 0, black 1.5rem, black calc(100% - 1.5rem), transparent 100%);
		}
	}

	&.stacked {
		.notifications {
			.notifications-inner {
				.notification {
					&:not(:first-child) {
						@apply pointer-events-none select-none;
					}
				}
			}
		}
	}
}

@keyframes notification-create {
	0% {
		opacity: 0;
	}
	100% {
		opacity: var(--notification-opacity);
	}
}
</style>
