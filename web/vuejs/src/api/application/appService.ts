import i18n from '@/errorhandling/i18n/'; // Stelle sicher, dass der Pfad zu deinem i18n-Ordner korrekt ist
import {ref} from 'vue'
import {useHttp} from "@/shared/http";

/*
export const errorMessage = ref<string>('')
export const additionalInformation = ref<string>('')
export const showAdditionalInformation = ref<boolean>( false);



interface CustomError {
    errorMessage: string;
    additionalInformation?:string;
}

 */


//TODO: Klasse anlegen

// TODO: hier die fetchApplication Funktion schreiben
// Funktion schreiben, die den HTTP Client verwendet
export async function fetchApplication(url: string):Promise<Response> {
    const client = useHttp()
    return await client.request(url)
}


//TODO: in die ui Schicht verschieben
/*
export async function updateErrorMessage(statusCode: string | undefined): Promise<void> {
    try {
        const error = i18n.errors[statusCode];
        const errorMessages: CustomError = {
            errorMessage: error.errorMessage,
            additionalInformation: error.additionalInformation
        };

        if (error) {
            throw errorMessages;
        }

    } catch (error) {
        errorMessage.value = error.errorMessage; // Hier wird die Fehlermeldung im errorMessage-Ref aktualisiert
        additionalInformation.value = error.additionalInformation;
    }
}


export async function toggleErrorInfo():Promise<void> {
    showAdditionalInformation.value = !showAdditionalInformation.value

}

 */
