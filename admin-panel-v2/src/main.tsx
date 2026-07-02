import '@/index.css';
import '@/global.css';
import '@/i18n/config';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from '@/utils/request';
import { RouterProvider } from 'react-router-dom';
import { router } from '@/App';
import { createRoot } from 'react-dom/client';

function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}

const rootElement = document.getElementById('root');
if (rootElement) {
  createRoot(rootElement).render(<Root />);
}
