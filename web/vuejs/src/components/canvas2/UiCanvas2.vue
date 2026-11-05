<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { AssetLoader } from '@/components/gostrict/assets';
import { type ProgramHandle, startProgram } from '@/components/gostrict/run';
import type { Session } from '@/components/gostrict/session';
import { frameCSS } from '@/components/shared/frame';
import { Canvas2 } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Canvas2;
}>();

const rootRef = ref<HTMLDivElement | null>(null);

const frameStyles = computed<string>(() => frameCSS(props.ui.frame).join(';'));

/**
 * Host is the object, which the bindings of package lib/canvas call. The program draws in CSS pixels, the host
 * scales the canvas to the pixels of the device.
 */
interface Host {
	context: CanvasRenderingContext2D;
	dpr: number;
	width: number;
	height: number;
	loadImage(url: string): unknown;
}

let handle: ProgramHandle | null = null;
let canvas: HTMLCanvasElement | null = null;
let assets: AssetLoader | null = null;
let starts = 0;

// start runs the program. A new version of the program takes over the state of the running one, if it has the same
// structure, unless fresh is set: then the program starts from the beginning, e.g. with new arguments.
async function start(fresh: boolean) {
	const previous = fresh ? null : (handle?.session.takeOver() ?? null);
	stop();
	const root = rootRef.value;
	if (!root || !props.ui.program?.bundle) return;
	const run = ++starts;

	const c = document.createElement('canvas');
	c.tabIndex = 0;
	c.className = 'block w-full h-full outline-hidden';
	root.appendChild(c);
	canvas = c;

	let session: Session | null = null;
	const loader = new AssetLoader((type, values) => session?.emit(type, values));
	assets = loader;
	const host: Host = {
		context: c.getContext('2d')!,
		dpr: 1,
		width: 0,
		height: 0,
		loadImage: (url: string) => loader.image(url),
	};

	const h = await startProgram({
		id: props.ui.id ?? '',
		program: props.ui.program,
		root,
		input: c,
		libs: {},
		host,
		resize(width, height, dpr) {
			host.width = width;
			host.height = height;
			host.dpr = dpr;
			c.width = Math.max(1, Math.round(width * dpr));
			c.height = Math.max(1, Math.round(height * dpr));
			host.context.setTransform(dpr, 0, 0, dpr, 0, 0);
		},
		// each frame starts with the scale of the device and no other transformation
		before() {
			host.context.setTransform(host.dpr, 0, 0, host.dpr, 0, 0);
		},
		render() {},
		attached(s) {
			session = s;
		},
		previous,
		alive: () => run === starts,
	});
	if (run !== starts) {
		h?.stop();
		return;
	}
	handle = h;
	// the inputs may have changed, while the program was loading
	h?.inputs(props.ui.program?.inputs?.value ?? []);
}

function stop() {
	starts++;
	handle?.stop();
	handle = null;
	assets?.dispose();
	assets = null;
	canvas?.remove();
	canvas = null;
}

onMounted(() => start(true));
onUnmounted(stop);

// Other arguments or another id start the program from the beginning, with the arguments and the channel of the
// id. A new version of the program, e.g. after an edit, replaces the running one and takes over its state, if the
// arguments and the id stay. Other inputs apply to the running program. A render of nago, which changes nothing of
// it, keeps the program and its state.
watch(
	[
		() => props.ui.id ?? '',
		() => props.ui.program?.args ?? '',
		() => `${props.ui.program?.bundle}|${props.ui.program?.init}`,
		() => (props.ui.program?.inputs?.value ?? []).join(','),
	],
	([id, args, version, inputs], [oldID, oldArgs, oldVersion, oldInputs]) => {
		if (id !== oldID || args !== oldArgs) {
			start(true);
		} else if (version !== oldVersion) {
			start(false);
		} else if (inputs !== oldInputs) {
			handle?.inputs(props.ui.program?.inputs?.value ?? []);
		}
	}
);
</script>

<template>
	<div :id="ui.id" ref="rootRef" :style="frameStyles" class="relative overflow-hidden"></div>
</template>
