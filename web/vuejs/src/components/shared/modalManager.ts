type ScrollEvent = WheelEvent | TouchEvent;

export class ModalManager {
	// containers which must stay interactive, even if a dialog is open
	static readonly keepInteractiveSelector = '#ora-notifications';

	private static openModalCount = 0;

	// stack of visible dialog containers, the last one is the topmost dialog and the only one which may hold the focus
	private static readonly dialogStack: HTMLElement[] = [];

	// elements which have been made inert by us and must be restored, once no dialog is open anymore
	private static readonly inertElements = new Set<HTMLElement>();

	private static lastTouch: { x: number; y: number } | undefined;

	static onModalOpen(): void {
		if (ModalManager.openModalCount === 0) {
			document.body.style.overflow = 'hidden';
			addEventListener('wheel', ModalManager.onScrollAttempt, { capture: true, passive: false });
			addEventListener('touchmove', ModalManager.onScrollAttempt, { capture: true, passive: false });
			addEventListener('touchstart', ModalManager.onTouchStart, { capture: true, passive: true });
		}
		ModalManager.openModalCount++;
	}

	static onModalClose(): void {
		if (ModalManager.openModalCount === 0) return;

		ModalManager.openModalCount--;
		if (ModalManager.openModalCount === 0) {
			document.body.style.overflow = 'visible';
			removeEventListener('wheel', ModalManager.onScrollAttempt, { capture: true });
			removeEventListener('touchmove', ModalManager.onScrollAttempt, { capture: true });
			removeEventListener('touchstart', ModalManager.onTouchStart, { capture: true });
			ModalManager.lastTouch = undefined;
		}
	}

	// registers the container as the topmost dialog. While it is the topmost dialog, the background is inert and the
	// tab order cycles within the container.
	static pushDialog(container: HTMLElement): void {
		ModalManager.removeFromStack(container);
		ModalManager.dialogStack.push(container);
		ModalManager.updateInert();
		ModalManager.updateFocusTrap();
	}

	static popDialog(container: HTMLElement): void {
		if (!ModalManager.removeFromStack(container)) return;

		ModalManager.updateInert();
		ModalManager.updateFocusTrap();
	}

	private static focusTrapActive = false;

	private static updateFocusTrap(): void {
		const active = ModalManager.dialogStack.length > 0;
		if (active === ModalManager.focusTrapActive) return;

		ModalManager.focusTrapActive = active;
		if (active) {
			document.addEventListener('keydown', ModalManager.onTabKeydown, true);
		} else {
			document.removeEventListener('keydown', ModalManager.onTabKeydown, true);
		}
	}

	private static readonly focusableSelector = [
		'a[href]',
		'area[href]',
		'button',
		'input:not([type="hidden"])',
		'select',
		'textarea',
		'iframe',
		'audio[controls]',
		'video[controls]',
		'summary',
		'[contenteditable]:not([contenteditable="false"])',
		'[tabindex]',
	].join(',');

	// returns all tabbable elements within the given root in document order.
	static getTabbableElements(root: Element): HTMLElement[] {
		return Array.from(root.querySelectorAll<HTMLElement>(ModalManager.focusableSelector)).filter(
			(el) =>
				el.tabIndex >= 0 &&
				!(el as HTMLButtonElement).disabled &&
				!el.closest('[inert]') &&
				el.getClientRects().length > 0
		);
	}

	// keeps the tab order cycling within the topmost dialog.
	private static onTabKeydown(e: KeyboardEvent): void {
		if (e.key !== 'Tab' || e.defaultPrevented) return;

		const container = ModalManager.topmostDialog();
		if (!container) return;

		const active = document.activeElement;
		if (active instanceof Element && active.closest(ModalManager.keepInteractiveSelector)) return;

		const tabbables = ModalManager.getTabbableElements(container);
		if (tabbables.length === 0) {
			e.preventDefault();
			return;
		}

		const first = tabbables[0];
		const last = tabbables[tabbables.length - 1];
		const insideDialog = active instanceof Node && container.contains(active);

		if (!insideDialog) {
			e.preventDefault();
			(e.shiftKey ? last : first).focus();
		} else if (e.shiftKey && active === first) {
			e.preventDefault();
			last.focus();
		} else if (!e.shiftKey && active === last) {
			e.preventDefault();
			first.focus();
		}
	}

	private static removeFromStack(container: HTMLElement): boolean {
		const idx = ModalManager.dialogStack.indexOf(container);
		if (idx < 0) return false;

		ModalManager.dialogStack.splice(idx, 1);
		return true;
	}

	// only the topmost dialog may hold the focus, all other dialogs are in the background.
	static isTopmostDialog(container: HTMLElement): boolean {
		return ModalManager.topmostDialog() === container;
	}

	private static topmostDialog(): HTMLElement | undefined {
		return ModalManager.dialogStack[ModalManager.dialogStack.length - 1];
	}

	// makes everything except the topmost dialog (and its ancestors) inert, so that neither the keyboard nor
	// assistive technologies can reach the background.
	private static updateInert(): void {
		ModalManager.inertElements.forEach((el) => (el.inert = false));
		ModalManager.inertElements.clear();

		const top = ModalManager.topmostDialog();
		if (!top) return;

		let current: HTMLElement = top;
		while (current.parentElement && current !== document.body) {
			for (const sibling of current.parentElement.children) {
				if (!(sibling instanceof HTMLElement) || sibling === current || sibling.inert) continue;
				if (sibling.matches(ModalManager.keepInteractiveSelector)) continue;
				if (['SCRIPT', 'STYLE', 'LINK', 'TEMPLATE'].includes(sibling.tagName)) continue;

				sibling.inert = true;
				ModalManager.inertElements.add(sibling);
			}

			current = current.parentElement;
		}
	}

	private static onTouchStart(e: TouchEvent): void {
		const t = e.touches[0];
		ModalManager.lastTouch = t ? { x: t.clientX, y: t.clientY } : undefined;
	}

	// prevents scrolling of the background, e.g. by the mouse wheel or touch gestures. Scrolling is only allowed, if
	// the event happens within the topmost dialog and an element between the event target and the dialog can still
	// scroll in the requested direction. Otherwise, the browser would chain the scroll to the background.
	private static onScrollAttempt(e: ScrollEvent): void {
		if (!e.cancelable) return;

		// keep pinch-to-zoom working
		if (e instanceof WheelEvent ? e.ctrlKey : e.touches.length > 1) return;

		const target = e.target;
		if (!(target instanceof Element)) {
			e.preventDefault();
			return;
		}

		if (target.closest(ModalManager.keepInteractiveSelector)) return;

		// without any dialog, e.g. for an overlay which disallows background scrolling, only the body is locked
		const boundary = ModalManager.topmostDialog() ?? document.body;
		if (!boundary.contains(target) || !ModalManager.canScrollWithin(target, boundary, e)) {
			e.preventDefault();
		}
	}

	private static scrollDelta(e: ScrollEvent): { dx: number; dy: number } {
		if (e instanceof WheelEvent) {
			return e.shiftKey && e.deltaX === 0 ? { dx: e.deltaY, dy: 0 } : { dx: e.deltaX, dy: e.deltaY };
		}

		const t = e.touches[0];
		if (!t) return { dx: 0, dy: 0 };

		const previous = ModalManager.lastTouch;
		ModalManager.lastTouch = { x: t.clientX, y: t.clientY };
		if (!previous) return { dx: 0, dy: 0 };

		// moving the finger up scrolls down
		return { dx: previous.x - t.clientX, dy: previous.y - t.clientY };
	}

	private static canScrollWithin(target: Element, boundary: HTMLElement, e: ScrollEvent): boolean {
		const { dx, dy } = ModalManager.scrollDelta(e);
		if (dx === 0 && dy === 0) return true;

		for (let el: Element | null = target; el && el !== boundary.parentElement; el = el.parentElement) {
			const style = getComputedStyle(el);

			if (dy !== 0 && /(auto|scroll|overlay)/.test(style.overflowY) && el.scrollHeight > el.clientHeight) {
				if (dy < 0 && el.scrollTop > 0) return true;
				if (dy > 0 && Math.ceil(el.scrollTop + el.clientHeight) < el.scrollHeight) return true;
			}

			if (dx !== 0 && /(auto|scroll|overlay)/.test(style.overflowX) && el.scrollWidth > el.clientWidth) {
				// scrollLeft is negative or zero in rtl layouts, thus compare absolute values
				const left = Math.abs(el.scrollLeft);
				if (dx < 0 && left > 0) return true;
				if (dx > 0 && Math.ceil(left + el.clientWidth) < el.scrollWidth) return true;
			}
		}

		return false;
	}
}
