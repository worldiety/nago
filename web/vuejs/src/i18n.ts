import { createI18n } from 'vue-i18n';

import de from '@/locales/de.json'
import en from '@/locales/en.json'

type MessageSchemaDe = typeof de;
type MessageSchemaEn = typeof en;


export default createI18n<[MessageSchemaDe | MessageSchemaEn]>({
    legacy: false,
    messages: {
        de: de,
        en: en,
    },
});