import type { SelectItem } from '@/shared/model/selectItem';

export interface FormField {
	type: 'TextField' | 'FileUploadField' | 'SelectField';
	label: string;
	id: string;
	value: string | null;
	hint: string;
	error: string;
	disabled: boolean;
	fileMultiple: boolean | null;
	fileAccept: string | null;
	selectMultiple: boolean | null;
	selectItems: SelectItem[];
	selectValues: string[];
}
