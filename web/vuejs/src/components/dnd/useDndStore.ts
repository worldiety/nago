import { ref } from 'vue';
import { defineStore } from 'pinia';


export const useDndStore = defineStore('dnd', () => {
	const dragging = ref(false);
	const forbidden = ref(false);
	const currentDragId = ref<string | null>(null);

	function startDrag(id: string) {
		dragging.value = true;
		forbidden.value = false;
		currentDragId.value = id;
	}

	function endDrag() {
		dragging.value = false;
		forbidden.value = false;
		currentDragId.value = null;
	}

	function enteredDropArea(allowed: boolean) {
		if (!dragging.value) return;
		forbidden.value = !allowed;
	}

	function leftDropArea() {
		if (!dragging.value) return;
		forbidden.value = false;
	}

	return { dragging, forbidden, currentDragId, startDrag, endDrag, enteredDropArea, leftDropArea };
});
