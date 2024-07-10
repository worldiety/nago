import {Length} from "@/shared/protocol/ora/length";

export function cssLengthValue(l?: Length): string {
	if (!l) {
		return ""
	}

	l = l.replaceAll("dp", "px")
	return l
}

export function cssLengthValue0Px(l?: Length): string {
	if (!l) {
		return "0px"
	}

	l = l.replaceAll("dp", "px")
	return l
}
