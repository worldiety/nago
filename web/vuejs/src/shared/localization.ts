import { activeLocale } from '@/i18n';

export function localizeNumber(rawNumber: number, options: Intl.NumberFormatOptions): string {
	return rawNumber.toLocaleString(activeLocale, options);
}
