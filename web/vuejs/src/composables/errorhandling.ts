import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

export interface CustomError {
	errorCode?: string
	message?: string
	additionalInformation?: string
}

export type ApplicationError = Error | unknown | Response | CustomError;

// by convention, composable function names start with "use"
export function useErrorHandling() {
	const error = ref<CustomError | null>(null);
	const i18n = useI18n();

	// a composable can update its managed state over time.
	function handleError(rawError: ApplicationError) {
		if (rawError instanceof Error) {
			console.log('rawError ist Error:');
			error.value = {
				message: 'TODO: Message definieren',
				additionalInformation: 'TODO: Message definieren',
			};
		} else if (rawError instanceof Response) {
			console.log('rawError ist Response');
			error.value = {
				message: String(i18n.t('httpErrorcodes.' + rawError.status + '.errorMessage')),
				additionalInformation: 'TODO: Message definieren',
			};
		} else if (rawError as CustomError) {
			console.log('rawError ist CustomError');
			const rawCustomError = rawError as CustomError;
			error.value = {
				message: String(i18n.t('customErrorcodes.' + rawCustomError.errorCode + '.errorMessage')),
				additionalInformation: String(i18n.t('customErrorcodes.' + rawCustomError.errorCode + '.additionalInformation')),
			};
		} else {
			console.log('rawError ist unknown');
			error.value = {
				message: String(rawError),
				additionalInformation: String(rawError),
			};
		}
	}

	return {
		error,
		handleError,
	};
}
