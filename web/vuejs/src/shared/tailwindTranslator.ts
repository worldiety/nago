export function gapSize2Tailwind(s: string): string {
	if (s == null || s == '') {
		return '';
	}

	if (s.endsWith('px') || s.endsWith('pt') || s.endsWith('rem')) {
		return 'gap-[' + s + ']';
	}

	return s;
}
