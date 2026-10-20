<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import * as THREE from 'three';
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js';
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { AssetLoader, sameOrigin } from '@/components/gostrict/assets';
import { type ProgramHandle, startProgram } from '@/components/gostrict/run';
import type { Session } from '@/components/gostrict/session';
import { frameCSS } from '@/components/shared/frame';
import { ThreeJS } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: ThreeJS;
}>();

const rootRef = ref<HTMLDivElement | null>(null);

const frameStyles = computed<string>(() => frameCSS(props.ui.frame).join(';'));

let handle: ProgramHandle | null = null;
let renderer: THREE.WebGLRenderer | null = null;
// canvas is created for each start, because a canvas, whose context has been released, gets no new one
let canvas: HTMLCanvasElement | null = null;
let shown: { scene: THREE.Scene; camera: THREE.Camera } | null = null;
let assets: AssetLoader | null = null;

function render() {
	if (renderer && shown) renderer.render(shown.scene, shown.camera);
}

async function start() {
	// a new version of the program takes over the state of the running one, if it has the same structure
	const previous = handle?.session.takeOver() ?? null;
	stop();
	const root = rootRef.value;
	if (!root || !props.ui.program?.bundle) return;

	const c = document.createElement('canvas');
	c.tabIndex = 0;
	c.className = 'block w-full h-full outline-hidden';
	c.style.touchAction = 'none';
	root.appendChild(c);
	canvas = c;

	const r = new THREE.WebGLRenderer({ canvas: c, antialias: props.ui.antialias ?? false });
	renderer = r;
	let session: Session | null = null;
	const loader = new AssetLoader((type, values) => session?.emit(type, values));
	assets = loader;

	// host is the object, which the bindings of package lib/three call
	const host = {
		width: 0,
		height: 0,
		show(scene: THREE.Scene, camera: THREE.Camera) {
			if (renderer === r) shown = { scene, camera };
		},
		shadows(on: boolean) {
			r.shadowMap.enabled = on;
		},
		loadTexture(url: string): THREE.Texture {
			const tex = new THREE.Texture();
			tex.colorSpace = THREE.SRGBColorSpace;
			// the bitmap is already upside down, as WebGL needs it
			tex.flipY = false;
			const img = loader.image(url, true);
			const apply = () => {
				tex.image = img.bitmap;
				tex.needsUpdate = true;
				if (!session?.animated) render();
			};
			if (img.bitmap) apply();
			else img.users.push(apply);
			return tex;
		},
		loadModel(url: string): THREE.Group {
			const group = new THREE.Group();
			try {
				new GLTFLoader().load(
					sameOrigin(url).href,
					(gltf) => {
						if (renderer !== r) return;
						group.add(gltf.scene);
						loader.loaded(url);
					},
					undefined,
					(e) => loader.failed(url, e)
				);
			} catch (e) {
				queueMicrotask(() => loader.failed(url, e));
			}
			return group;
		},
	};

	const h = await startProgram({
		id: props.ui.id ?? '',
		program: props.ui.program,
		root,
		input: c,
		libs: { THREE },
		host,
		resize(width, height, dpr) {
			host.width = width;
			host.height = height;
			r.setPixelRatio(dpr);
			r.setSize(width, height, false);
		},
		render,
		attached(s) {
			session = s;
		},
		previous,
	});
	if (renderer !== r) {
		// unmounted or restarted while the program was loading, which has released the renderer
		h?.stop();
		return;
	}
	handle = h;
}

// dispose releases the GPU resources of the scene, because a browser keeps only a few WebGL contexts
function dispose(scene: THREE.Scene) {
	scene.traverse((o) => {
		const mesh = o as THREE.Mesh;
		mesh.geometry?.dispose();
		const materials = Array.isArray(mesh.material) ? mesh.material : mesh.material ? [mesh.material] : [];
		for (const m of materials) {
			for (const v of Object.values(m)) {
				if (v instanceof THREE.Texture) v.dispose();
			}
			m.dispose();
		}
	});
	if (scene.background instanceof THREE.Texture) scene.background.dispose();
}

function stop() {
	handle?.stop();
	handle = null;
	if (shown) dispose(shown.scene);
	shown = null;
	assets?.dispose();
	assets = null;
	if (renderer) {
		renderer.dispose();
		renderer.forceContextLoss();
	}
	renderer = null;
	canvas?.remove();
	canvas = null;
}

onMounted(start);
onUnmounted(stop);

// a new version of the program, e.g. after an edit, replaces the running one, while a render of nago, which
// changes nothing of it, keeps the program and its state
watch(
	() => `${props.ui.program?.bundle}|${props.ui.program?.init}|${props.ui.antialias}`,
	() => start()
);
</script>

<template>
	<div :id="ui.id" ref="rootRef" :style="frameStyles" class="relative overflow-hidden"></div>
</template>
