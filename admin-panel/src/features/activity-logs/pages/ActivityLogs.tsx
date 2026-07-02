import { useState } from 'react';
import { Button, Modal } from 'antd';
import { ProDescriptions } from '@ant-design/pro-components';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { Link } from 'react-router';
import { useTranslation } from 'react-i18next';
import { PageContainer } from '@/components/PageContainer';
import { useActivityLogActions } from '@/api/admin.hooks';
import { fetchActivityLogs } from '@/api/admin';
import type { AdminActivityLog } from '@/api/admin';
import { formatISODate } from '@/utils/format';
import { EyeOutlined } from '@ant-design/icons';

export default function ActivityLogsPage() {
  const { t, i18n } = useTranslation();
  const [detailLog, setDetailLog] = useState<AdminActivityLog | null>(null);
  const { data: actions } = useActivityLogActions(i18n.language);

  const columns: ProColumns<AdminActivityLog>[] = [
    { title: t('activityLogs.id'), dataIndex: 'id', width: 60, hideInSearch: true },
    {
      title: t('activityLogs.user'),
      dataIndex: 'username',
      render: (_, r) => <Link to={`/admin/users/${r.userId}`}>{r.username}</Link>,
      hideInSearch: true,
    },
    {
      title: t('activityLogs.action'),
      dataIndex: 'action',
      valueType: 'select',
      fieldProps: { options: (actions ?? []).map((a) => ({ label: a.label, value: a.action })) },
      render: (_, r) => r.actionLabel,
    },
    {
      title: t('activityLogs.resource'),
      render: (_, r) => `${r.resourceType}: ${r.resourceName}`,
      hideInSearch: true,
    },
    {
      title: t('activityLogs.ip'),
      dataIndex: 'ip',
    },
    { title: t('activityLogs.os'), dataIndex: 'os', hideInSearch: true },
    { title: t('activityLogs.browser'), dataIndex: 'browser', hideInSearch: true, ellipsis: true },
    {
      title: t('activityLogs.time'),
      dataIndex: 'createdAt',
      render: (_, record) => formatISODate(record.createdAt),
      hideInSearch: true,
      sorter: true,
      defaultSortOrder: 'descend',
    },
    {
      title: '',
      width: 80,
      hideInSearch: true,
      render: (_, r) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailLog(r)}>
          {t('activityLogs.detail')}
        </Button>
      ),
    },
  ];

  return (
    <PageContainer title={t('activityLogs.title')}>
      <ProTable<AdminActivityLog>
        headerTitle={t("activityLogs.title")}
        rowKey="id"
        columns={columns}
        request={async (params, sorter) => {
          const sortField = Object.keys(sorter)[0];
          const sortOrder = sortField
            ? (sorter as Record<string, string>)[sortField] === 'ascend'
              ? 'asc'
              : 'desc'
            : undefined;
          const uid = params.userId ? Number(params.userId) : undefined;
          const res = await fetchActivityLogs({
            limit: params.pageSize!,
            offset: (params.current! - 1) * params.pageSize!,
            user_id: uid && !Number.isNaN(uid) ? uid : undefined,
            action: (params.action as string) || undefined,
            ip: (params.ip as string) || undefined,
            sortBy: sortField as string | undefined,
            sortOrder,
          });
          return { data: res.items, success: true, total: res.total };
        }}
        search={{
          labelWidth: 'auto',
          defaultCollapsed: false,
        }}
        pagination={{
          showSizeChanger: true,
          showTotal: (total) => t('activityLogs.total_0', { count: total }),
        }}
        size="small"
      />

      <Modal
        title={t('activityLogs.detailTitle')}
        open={!!detailLog}
        onCancel={() => setDetailLog(null)}
        footer={null}
        width={600}
      >
        {detailLog && (
          <ProDescriptions column={1} size="small">
            <ProDescriptions.Item label={t('activityLogs.id')}>{detailLog.id}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.user')}>
              {detailLog.username} ({detailLog.userId})
            </ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.action')}>{detailLog.actionLabel}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.resource')}>
              {detailLog.resourceType}: {detailLog.resourceName}
            </ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.ip')}>{detailLog.ip || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.region')}>{detailLog.ipRegion || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.userAgent')}>{detailLog.userAgent || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.os')}>{detailLog.os || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.browser')}>{detailLog.browser || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.extra')}>
              {detailLog.extra ? JSON.stringify(detailLog.extra, null, 2) : '-'}
            </ProDescriptions.Item>
            <ProDescriptions.Item label={t('activityLogs.time')}>
              {formatISODate(detailLog.createdAt)}
            </ProDescriptions.Item>
          </ProDescriptions>
        )}
      </Modal>
    </PageContainer>
  );
}