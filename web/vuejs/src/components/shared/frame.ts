import {Frame} from "@/shared/protocol/ora/frame";

export function createFrameStyles(frame?: Frame):string {
	const styles: string[] = [];
	if (frame?.width) {
		styles.push("width:"+frame.width.replaceAll("dp", "px"))
	}

	if (frame?.minWidth) {
		styles.push("min-width:"+frame.minWidth.replaceAll("dp", "px"))
	}

	if (frame?.maxWidth) {
		styles.push("max-width:"+frame.maxWidth.replaceAll("dp", "px"))
	}

	if (frame?.height) {
		styles.push("height:"+frame.height.replaceAll("dp", "px"))
	}

	if (frame?.minHeight) {
		styles.push("min-height:"+frame.minHeight.replaceAll("dp", "px"))
	}

	if (frame?.maxHeight) {
		styles.push("max-height:"+frame.maxHeight.replaceAll("dp", "px"))
	}

	return styles.join('; ');
}
