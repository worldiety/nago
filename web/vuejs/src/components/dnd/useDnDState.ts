import { ref } from 'vue';

const currentDragId = ref<string | null>(null);

export function useDnDState() {
	function startDrag(id: string) {
		currentDragId.value = id;
	}

	function endDrag() {
		currentDragId.value = null;
	}

	return { currentDragId, startDrag, endDrag };
}
