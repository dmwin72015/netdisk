import { Table, Result, Empty } from 'antd';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import type { ListQueryResult } from '../../types/query';
import { useTranslation } from 'react-i18next';

interface ProTableProps<T extends object> {
  queryResult: ListQueryResult<T>;
  columns: ColumnsType<T>;
  rowKey: string | ((record: T) => string);
  headerTitle?: string;
  defaultPageSize?: number;
  showSizeChanger?: boolean;
  pagination?: boolean | TablePaginationConfig;
}

export default function ProTable<T extends object>({
  queryResult,
  columns,
  rowKey,
  headerTitle,
  defaultPageSize = 20,
  showSizeChanger = true,
  pagination: paginationProp,
}: ProTableProps<T>) {
  const { t } = useTranslation();

  const { data, total, isLoading, isFetching, isError, error, refetch } = queryResult;

  if (isError) {
    return (
      <Result
        status="error"
        title="加载失败"
        subTitle={error?.message}
        extra={
          // eslint-disable-next-line jsx-a11y/anchor-is-valid
          <a onClick={() => refetch()} style={{ cursor: 'pointer' }}>
            重试
          </a>
        }
      />
    );
  }

  if (!isLoading && data.length === 0) {
    return <Empty description={t('common.noData')} style={{ padding: 40 }} />;
  }

  const pagination: TablePaginationConfig | false =
    paginationProp === false
      ? false
      : {
          ...paginationProp,
          current: (paginationProp as TablePaginationConfig)?.current ?? 1,
          pageSize: (paginationProp as TablePaginationConfig)?.pageSize ?? defaultPageSize,
          total,
          showSizeChanger,
          showTotal: (tot: number) => t('common.total', { count: tot }),
        };

  return (
    <Table<T>
      rowKey={rowKey}
      columns={columns}
      dataSource={data}
      loading={isLoading || isFetching}
      pagination={pagination}
      size="small"
      title={
        headerTitle
          ? () => headerTitle
          : undefined
      }
    />
  );
}
