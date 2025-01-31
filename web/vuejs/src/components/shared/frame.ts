import { cssLengthValue } from '@/components/shared/length';
import {Frame} from "@/shared/proto/nprotoc_gen";

export function frameCSS(frame?: Frame): string[] {
	const styles: string[] = [];
	if (!frame) {
		return styles;
	}

	if (!frame.width.isZero()) {
		styles.push('width:' + cssLengthValue(frame.width.value));
	}

	if (!frame.minWidth.isZero()) {
		styles.push('min-width:' + cssLengthValue(frame.minWidth.value));
	}

	if (!frame.maxWidth.isZero()) {
		styles.push('max-width:' + cssLengthValue(frame.maxWidth.value));
	}

	if (!frame.height.isZero()) {
		styles.push('height:' + cssLengthValue(frame.height.value));
	}

	if (!frame.minHeight.isZero()) {
		styles.push('min-height:' + cssLengthValue(frame.minHeight.value));
	}

	if (!frame.maxHeight) {
		styles.push('max-height:' + cssLengthValue(frame.maxHeight));
	}

	return styles;
}
