import type { Chart as ChartProto } from '@/shared/proto/nprotoc_gen';
import { RoundingTypeValues } from '@/shared/proto/nprotoc_gen';

export class Chart {
	public static DataLabelFormatter(
		chart?: ChartProto
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
	): (val: string | number | number[], opts?: any) => string | number | (string | number)[] {
		return (val: string | number | number[]): string | number | (string | number)[] => {
			if (!chart || typeof val !== 'number') return val;

			switch (chart.labelRounding) {
				case RoundingTypeValues.Round:
					return Math.round(val);
				case RoundingTypeValues.Floor:
					return Math.floor(val);
				case RoundingTypeValues.Ceiling:
					return Math.ceil(val);
			}

			return val;
		};
	}
}
