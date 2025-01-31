import {cssLengthValue} from '@/components/shared/length';
import {Padding} from "@/shared/proto/nprotoc_gen";

// paddingCSS applies the padding length values. Note, that negative paddings are interpreted as negative margins,
// because negative padding values are not allowed and it seems practical to move views around for some nice effects.
export function paddingCSS(padding: Padding): string[] {
	const styles: string[] = [];

	if (!padding) {
		return styles;
	}

	if (!padding.bottom.isZero()) {
		if (padding.bottom.value.startsWith('-')) {
			styles.push(`margin-bottom:${cssLengthValue(padding.bottom.value)}`);
		} else {
			styles.push(`padding-bottom:${cssLengthValue(padding.bottom.value)}`);
		}
	}

	if (!padding.top.isZero()) {
		if (padding.top.value.startsWith('-')) {
			styles.push(`margin-top: ${cssLengthValue(padding.top.value)}`);
		} else {
			styles.push(`padding-top:${cssLengthValue(padding.top.value)}`);
		}
	}

	if (!padding.right.isZero()) {
		if (padding.right.value.startsWith('-')) {
			styles.push(`margin-right:${cssLengthValue(padding.right.value)}`);
		} else {
			styles.push(`padding-right:${cssLengthValue(padding.right.value)}`);
		}
	}

	if (!padding.left.isZero()) {
		if (padding.left.value.startsWith('-')) {
			styles.push(`margin-left:${cssLengthValue(padding.left.value)}`);
		} else {
			styles.push(`padding-left:${cssLengthValue(padding.left.value)}`);
		}
	}

	return styles;
}

// marginCSS is like padding but interprets all padding lengths as margin length
export function marginCSS(padding?: Padding): string[] {
	const styles: string[] = [];

	if (!padding) {
		return styles;
	}

	if (!padding.bottom.isZero()) {
		styles.push(`margin-bottom:${cssLengthValue(padding.bottom.value)}`);
	}

	if (!padding.top.isZero()) {
		styles.push(`margin-top: ${cssLengthValue(padding.top.value)}`);
	}

	if (!padding.right.isZero()) {
		styles.push(`margin-right:${cssLengthValue(padding.right.value)}`);
	}

	if (!padding.left.isZero()) {
		styles.push(`margin-left:${cssLengthValue(padding.left.value)}`);
	}

	return styles;
}
