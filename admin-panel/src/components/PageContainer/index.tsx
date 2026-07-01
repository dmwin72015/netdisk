import type { ReactNode } from 'react';
import { Space } from 'antd';

interface PageContainerProps {
  title?: string;
  extra?: ReactNode;
  children: ReactNode;
}

function PageContainer({ title, extra, children }: PageContainerProps) {
  return (
    <div>
      {(title || extra) && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 24,
          }}
        >
          {title && <h2 style={{ margin: 0 }}>{title}</h2>}
          {extra && <Space>{extra}</Space>}
        </div>
      )}
      {children}
    </div>
  );
}

export { PageContainer };
export default PageContainer;
export type { PageContainerProps };
