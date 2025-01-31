import { cssLengthValue } from '@/components/shared/length';
import {
	Position,
	PositionAbsolute,
	PositionDefault,
	PositionFixed,
	PositionOffset,
	PositionSticky,
} from '@/shared/protocol/ora/position';

export function positionCSS(position?: Position): string[] {
	const styles: string[] = [];

	if (!position) {
		return styles;
	}

	switch (position.k) {
		case PositionDefault:
			styles.push('position:static');
			break;
		case PositionAbsolute:
			styles.push('position:absolute');
			break;
		case PositionOffset:
			styles.push('position:relative');
			break;
		case PositionSticky:
			styles.push('position:sticky');
			break;
		case PositionFixed:
			styles.push('position:fixed');
			break;
	}

	if (position.l) {
		styles.push('left:' + cssLengthValue(position.l));
	}

	if (position.t) {
		styles.push('top:' + cssLengthValue(position.t));
	}

	if (position.r) {
		styles.push('right:' + cssLengthValue(position.r));
	}

	if (position.b) {
		styles.push('bottom:' + cssLengthValue(position.b));
	}

	console.log('fuck', styles, position);

	return styles;
}
