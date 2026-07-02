import { useTranslation } from 'react-i18next';

export function useRequestErrorConfig() {
  const { t } = useTranslation();

  return {
    errorHandler: (error: Error) => {
      if (!navigator.onLine) {
        return t('app.request.offline');
      }
      return error.message;
    },
  };
}
