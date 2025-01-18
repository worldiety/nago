import {Frame} from "@/shared/protocol/ora/frame";
import {cssLengthValue} from "@/components/shared/length";

export function frameCSS(frame?: Frame): string[] {
	const styles: string[] = [];
	if (!frame) {
		return styles;
	}
	
	if (frame?.w) {
		styles.push("width:" + cssLengthValue(frame.w))
	}

	if (frame?.wi) {
		styles.push("min-width:" + cssLengthValue(frame.wi))
	}

	if (frame?.wx) {
		styles.push("max-width:" + cssLengthValue(frame.wx))
	}

	if (frame?.h) {
		styles.push("height:" + cssLengthValue(frame.h))
	}

	if (frame?.hi) {
		styles.push("min-height:" + cssLengthValue(frame.hi))
	}

	if (frame?.hx) {
		styles.push("max-height:" + cssLengthValue(frame.hx))
	}

	return styles
}
