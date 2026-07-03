import type { ReactNode } from "react";
import { PageContainer as ProPageContainer } from "@ant-design/pro-layout";

interface PageContainerProps {
  title?: string;
  extra?: ReactNode;
  children: ReactNode;
}

/**
 * Wraps @ant-design/pro-layout's PageContainer.
 * ProTable and other components handle their own white card background.
 */
function PageContainer({ title, extra, children }: PageContainerProps) {
  return (
    <ProPageContainer
      title={title}
      extra={extra}
      token={{
        paddingInlinePageContainerContent: 24,
      }}
    >
      {children}
    </ProPageContainer>
  );
}

export { PageContainer };
export default PageContainer;
export type { PageContainerProps };
