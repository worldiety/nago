import {cssLengthValue} from "@/components/shared/length";
import {Padding} from "@/shared/protocol/ora/padding";


export function createPaddingStyles(frame?: Padding): string {
	if (frame?.r && frame?.t && frame?.b && frame?.l && frame.r===frame.l && frame.t===frame.b && frame.t===frame.r){
			return `padding: ${cssLengthValue(frame.b)};`
	}

	const styles: string[] = [];
	if (frame?.b) {
		styles.push("padding-bottom:" + cssLengthValue(frame.b))
	}

	if (frame?.t) {
		styles.push("padding-top:" + cssLengthValue(frame.t))
	}

	if (frame?.r) {
		styles.push("padding-right:" + cssLengthValue(frame.r))
	}

	if (frame?.l) {
		styles.push("padding-left:" + cssLengthValue(frame.l))
	}


	return styles.join('; ');
}
