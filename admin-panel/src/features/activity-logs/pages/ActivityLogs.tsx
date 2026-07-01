import { useCallback, useState } from 'react';
import { Button, Modal, Descriptions, Select, Input } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { PageContainer } from '../../../components/PageContainer';
import { SearchForm } from '../../../components/SearchForm';
import type { SearchField } from '../../../components/SearchForm';
import { ProTable } from '../../../components/ProTable';
import { useTableUrlState } from '../../../hooks/useTableUrlState';
import { useActivityLogs, useActivityLogActions } from '../../../api/admin.hooks';
import type { AdminActivityLog } from '../../../api/admin';
import { formatISODate } from '../../../utils/format';

export default function ActivityLogsPage() {
  const { t, i18n } = useTranslation();
  const [detailLog, setDetailLog] = useState<AdminActivityLog | null>(null);
  const { data: actions } = useActivityLogActions(i18n.language);

  const { params, setParams, resetParams } = useTableUrlState({
    page: 1,
    pageSize: 20,
    action: '',
    ip: '',
  });

  const queryResult = useActivityLogs({
    limit: params.pageSize,
    offset: (params.page - 1) * params.pageSize,
    action: params.action || undefined,
    ip: params.ip || undefined,
  });

  const handleSearch = useCallback(
    (values: Record<string, unknown>) => {
      setParams({
        page: 1,
        action: (values.action as string) || '',
        ip: (values.ip as string) || '',
      });
    },
    [setParams],
  );

  const handlePageChange = useCallback(
    (page: number, pageSize: number) => {
      setParams({ page, pageSize });
    },
    [setParams],
  );

  const columns: ColumnsType<AdminActivityLog> = [
    { title: t('activityLogs.id'), dataIndex: 'id', width: 60 },
    {
      title: t('activityLogs.user'),
      dataIndex: 'username',
      render: (_, r) => <Link to={`/admin/users/${r.userId}`}>{r.username}</Link>,
    },
    {
      title: t('activityLogs.action'),
      dataIndex: 'actionLabel',
      width: 160,
    },
    {
      title: t('activityLogs.resource'),
      render: (_, r) => `${r.resourceType}: ${r.resourceName}`,
    },
    {
      title: t('activityLogs.ip'),
      dataIndex: 'ip',
      width: 140,
    },
    { title: t('activityLogs.os'), dataIndex: 'os', width: 100 },
    { title: t('activityLogs.browser'), dataIndex: 'browser', ellipsis: true },
    {
      title: t('activityLogs.time'),
      dataIndex: 'createdAt',
      render: (v: string) => formatISODate(v),
      width: 170,
    },
    {
      title: '',
      width: 80,
      render: (_, r) => (
        <Button size="small" onClick={() => setDetailLog(r)}>
          {t('activityLogs.detail')}
        </Button>
      ),
    },
  ];

  const searchFields: SearchField[] = [
    {
      name: 'action',
      label: t('activityLogs.action'),
      children: (
        <Select
          allowClear
          placeholder={t('activityLogs.selectAction')}
          options={(actions ?? []).map((a) => ({ label: a.label, value: a.action }))}
          style={{ width: 200 }}
        />
      ),
    },
    {
      name: 'ip',
      label: t('activityLogs.ip'),
      children: <Input placeholder={t('activityLogs.ipPlaceholder')} allowClear />,
    },
  ];

  return (
    <PageContainer title={t('activityLogs.title')}>
      <SearchForm
        fields={searchFields}
        onSearch={handleSearch}
        onReset={resetParams}
        initialValues={{
          action: params.action || undefined,
          ip: params.ip || undefined,
        }}
      />
      <ProTable<AdminActivityLog>
        rowKey="id"
        columns={columns}
        queryResult={queryResult}
        pagination={{ current: params.page, pageSize: params.pageSize }}
        onChange={(pag) => {
          if (pag.current && pag.pageSize) {
            handlePageChange(pag.current, pag.pageSize);
          }
        }}
      />

      <Modal
        title={t('activityLogs.detailTitle')}
        open={!!detailLog}
        onCancel={() => setDetailLog(null)}
        footer={null}
        width={600}
      >
        {detailLog && (
          <Descriptions column={1} size="small">
            <Descriptions.Item label={t('activityLogs.id')}>{detailLog.id}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.user')}>
              {detailLog.username} ({detailLog.userId})
            </Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.action')}>{detailLog.actionLabel}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.resource')}>
              {detailLog.resourceType}: {detailLog.resourceName}
            </Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.ip')}>{detailLog.ip || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.region')}>{detailLog.ipRegion || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.userAgent')}>{detailLog.userAgent || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.os')}>{detailLog.os || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.browser')}>{detailLog.browser || '-'}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.extra')}>
              {detailLog.extra ? JSON.stringify(detailLog.extra, null, 2) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.time')}>
              {formatISODate(detailLog.createdAt)}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </PageContainer>
  );
}