import { PageContainer as ProPageContainer } from '@ant-design/pro-layout';
import type { ReactNode } from 'react';

interface PageContainerProps {
  title?: string;
  extra?: ReactNode;
  children: ReactNode;
}

/**
 * Wraps @ant-design/pro-layout's PageContainer with a white card content area.
 * This matches Ant Design Pro's page layout pattern.
 */
function PageContainer({ title, extra, children }: PageContainerProps) {
  return (
    <ProPageContainer title={title} extra={extra}>
      <div style={{ background: '#fff', padding: 24, borderRadius: 8 }}>
        {children}
      </div>
    </ProPageContainer>
  );
}

export { PageContainer };
export default PageContainer;
export type { PageContainerProps };