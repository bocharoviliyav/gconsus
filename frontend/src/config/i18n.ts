import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { config, STORAGE_KEYS } from './index';

import enTranslation from '../locales/en/translation.json';
import ruTranslation from '../locales/ru/translation.json';

const resources = {
  en: {
    translation: enTranslation,
  },
  ru: {
    translation: ruTranslation,
  },
};

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: localStorage.getItem(STORAGE_KEYS.locale) || config.app.defaultLocale,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // React already escapes
    },
    react: {
      useSuspense: false,
    },
  });

export default i18n;
