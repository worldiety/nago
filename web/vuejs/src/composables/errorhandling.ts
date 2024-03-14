import { ref } from 'vue'

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
                message: 'TODO: Message definieren',
                title: 'TODO: Message definieren',
            }
        } else if (rawError instanceof Response) {
            console.log('rawError ist Response')
            error.value = {
                message: 'TODO: Message definieren',
                title: 'TODO: Message definieren',
            }
        } else if (typeof (rawError as CustomError).title === 'string' && typeof (rawError as CustomError).message === 'string')  {
            const rawCustomError = rawError as CustomError

            error.value = {
                message: rawCustomError.message,
                title: rawCustomError.title,
                additionalInformation: rawCustomError.additionalInformation,
            }
        } else {
            console.log('rawError ist unknown')
            error.value = {
                message: String(rawError),
                title: 'Unbekannter Fehler',
            }
        }
    }


    return {
        error,
        handleError,
    }
}