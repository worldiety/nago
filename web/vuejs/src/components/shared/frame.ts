import { cssLengthValue } from '@/components/shared/length';
import { Frame } from '@/shared/proto/nprotoc_gen';

export function frameCSS(frame?: Frame): string[] {
	const styles: string[] = [];
	if (!frame) {
		return styles;
	}

	if (frame.width) {
		styles.push('width:' + cssLengthValue(frame.width));
	}

	if (frame.minWidth) {
		styles.push('min-width:' + cssLengthValue(frame.minWidth));
	}

	if (frame.maxWidth) {
		styles.push('max-width:' + cssLengthValue(frame.maxWidth));
	}

	if (frame.height) {
		styles.push('height:' + cssLengthValue(frame.height));
	}

	if (frame.minHeight) {
		styles.push('min-height:' + cssLengthValue(frame.minHeight));
	}

	if (frame.maxHeight) {
		styles.push('max-height:' + cssLengthValue(frame.maxHeight));
	}

	return styles;
}
