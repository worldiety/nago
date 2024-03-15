import { ref } from 'vue'
import {translate} from "@/i18n";


export interface CustomError {
    title: string;
    message: string;
    additionalInformation?:string;
}


// by convention, composable function names start with "use"
export function useErrorHandling() {
    const error = ref<CustomError | null>(null)

    // a composable can update its managed state over time.
    function handleError(rawError: Error | unknown | Response | CustomError) {
        console.log(rawError)

        if(rawError instanceof Error) {
            console.log('rawError ist Error:')
            error.value = {
                title: 'TODO: Message definieren',
                message: translate('httpErrorcodes.404.errorMessage'),
                additionalInformation: 'TODO: Message definieren'
            }
        } else if (rawError instanceof Response) {
            console.log('rawError ist Response')
            error.value = {
                title: 'TODO: Message definieren',
                message: 'TODO: Message definieren',
                additionalInformation: 'TODO: Message definieren'
            }
        } else if (typeof (rawError as CustomError).title === 'string' && typeof (rawError as CustomError).message === 'string')  {
            const rawCustomError = rawError as CustomError
            error.value = {
                title: rawCustomError.title,
                message: rawCustomError.message,
                additionalInformation: rawCustomError.additionalInformation,
            }
        } else {
            console.log('rawError ist unknown')
            error.value = {
                title: 'Unbekannter Fehler',
                message: String(rawError),
                additionalInformation: String(rawError),
            }
        }
    }


    return {
        error,
        handleError,
    }
}