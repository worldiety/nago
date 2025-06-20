import { TextFieldStyleValues } from '@/shared/proto/nprotoc_gen';

export enum InputWrapperStyle {
	REDUCED,
	BASIC,
}

export function inputWrapperStyleFrom(textFieldStyle?: TextFieldStyleValues): InputWrapperStyle | undefined {
	if (textFieldStyle === TextFieldStyleValues.TextFieldReduced) {
		return InputWrapperStyle.REDUCED;
	}
	if (textFieldStyle === TextFieldStyleValues.TextFieldBasic) {
		return InputWrapperStyle.BASIC;
	}
}
