let openModalCount = 0;

export function onModalOpen() {
	if (openModalCount === 0) {
		document.body.style.overflow = 'hidden';
	}
	openModalCount++;
}

export function onModalClose() {
	openModalCount = Math.max(0, openModalCount - 1);
	if (openModalCount === 0) {
		document.body.style.overflow = 'visible';
	}
}
