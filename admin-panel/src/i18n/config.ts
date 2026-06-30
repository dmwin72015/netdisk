import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import zhCN from './locales/zh-CN';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: { zh: { translation: zhCN } },
    fallbackLng: 'zh',
    interpolation: { escapeValue: false },
  });

export default i18n;