import { cssLengthValue } from '@/components/shared/length';
import { Position, PositionTypeValues } from '@/shared/proto/nprotoc_gen';

export function positionCSS(position?: Position): string[] {
	const styles: string[] = [];

	if (!position) {
		return styles;
	}

	switch (position.kind.value) {
		case PositionTypeValues.PositionDefault:
			//styles.push('position:static'); // TODO not sure if we should switch that to change inherit behavior
			break;
		case PositionTypeValues.PositionAbsolute:
			styles.push('position:absolute');
			break;
		case PositionTypeValues.PositionOffset:
			styles.push('position:relative');
			break;
		case PositionTypeValues.PositionSticky:
			styles.push('position:sticky');
			break;
		case PositionTypeValues.PositionFixed:
			styles.push('position:fixed');
			break;
	}

	if (!position.left.isZero()) {
		styles.push('left:' + cssLengthValue(position.left.value));
	}

	if (!position.top.isZero()) {
		styles.push('top:' + cssLengthValue(position.top.value));
	}

	if (!position.right.isZero()) {
		styles.push('right:' + cssLengthValue(position.right.value));
	}

	if (!position.bottom.isZero()) {
		styles.push('bottom:' + cssLengthValue(position.bottom.value));
	}

	return styles;
}
