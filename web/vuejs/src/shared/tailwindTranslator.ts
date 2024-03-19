export function textColor2Tailwind(s: string): string {
	if (s == null || s == '') {
		return '';
	}

	if (s.startsWith('#')) {
		return 'text-[' + s + ']';
	}

	return s;
}

export function textSize2Tailwind(s: string): string {
	if (s == null || s == '') {
		return '';
	}

	if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
		return 'text-[' + s + ']';
	}

	return 'text-' + s;
}

export function gapSize2Tailwind(s: string): string {
	if (s == null || s == '') {
		return '';
	}

	if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
		return 'gap-[' + s + ']';
	}

	return s;
}
