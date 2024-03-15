import { createI18n } from 'vue-i18n';

import de from '@/locales/de.json'
import en from '@/locales/en.json'

type MessageSchemaDe = typeof de;
type MessageSchemaEn = typeof en;


const i18n =  createI18n<[MessageSchemaDe | MessageSchemaEn], 'de' | 'en'>({
    legacy: false, // set `false`, to use Composition API
    locale: navigator.language,
    fallbackLocale: 'de',
    messages: {
        'de': de,
        'en': en,
    },
});

export default i18n


export function translate(key: string): string {
    const { t } = i18n.global;
    return t(key);
}