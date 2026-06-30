import { useState } from 'react';
import { Table, Input, Select, Button, Space, Modal, Descriptions, DatePicker, Result } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useActivityLogs, useActivityLogActions } from '../api/admin.hooks';
import type { ActivityLogParams, AdminActivityLog } from '../api/admin';
import type { ColumnsType } from 'antd/es/table';
import type { Dayjs } from 'dayjs';

const { RangePicker } = DatePicker;

export default function ActivityLogs() {
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [userId, setUserId] = useState<string | undefined>();
  const [action, setAction] = useState<string | undefined>();
  const [ip, setIp] = useState('');
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null);
  const [detailLog, setDetailLog] = useState<AdminActivityLog | null>(null);

  const params: ActivityLogParams = { limit: pageSize, offset: (page - 1) * pageSize };
  const uid = userId ? Number(userId) : undefined;
  if (uid && !Number.isNaN(uid)) params.user_id = uid;
  if (action) params.action = action;
  if (ip) params.ip = ip;
  if (dateRange && dateRange[0] && dateRange[1]) {
    params.created_from = dateRange[0].format('YYYY-MM-DD');
    params.created_to = dateRange[1].format('YYYY-MM-DD');
  }

  const { data, isLoading, error } = useActivityLogs(params);
  const { data: actions } = useActivityLogActions();

  const logs = data?.items ?? [];
  const total = data?.total ?? 0;

  const columns: ColumnsType<AdminActivityLog> = [
    { title: t('activityLogs.id'), dataIndex: 'id', width: 60 },
    {
      title: t('activityLogs.user'),
      dataIndex: 'username',
      render: (_, r) => <Link to={`/admin/users/${r.userId}`}>{r.username}</Link>,
    },
    { title: t('activityLogs.action'), dataIndex: 'actionLabel' },
    {
      title: t('activityLogs.resource'),
      render: (_, r) => `${r.resourceType}: ${r.resourceName}`,
    },
    { title: t('activityLogs.ip'), dataIndex: 'ip' },
    { title: t('activityLogs.os'), dataIndex: 'os' },
    { title: t('activityLogs.browser'), dataIndex: 'browser' },
    {
      title: t('activityLogs.time'),
      dataIndex: 'createdAt',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString(),
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

  const handleFilter = () => {
    setPage(1);
  };

  return (
    <div>
      <h2 style={{ marginBottom: 16 }}>{t('activityLogs.title')}</h2>

      <Space wrap style={{ marginBottom: 16 }}>
        <Input
          placeholder={t('activityLogs.userId')}
          type="number"
          value={userId ?? ''}
          onChange={(e) => setUserId(e.target.value || undefined)}
          style={{ width: 120 }}
          allowClear
          onClear={() => { setUserId(undefined); setPage(1); }}
        />
        <Select
          placeholder={t('activityLogs.selectAction')}
          allowClear
          value={action}
          onChange={(val) => { setAction(val); setPage(1); }}
          style={{ width: 180 }}
          options={(actions ?? []).map((a) => ({ label: a.label, value: a.action }))}
        />
        <Input
          placeholder={t('activityLogs.ipPlaceholder')}
          value={ip}
          onChange={(e) => setIp(e.target.value)}
          style={{ width: 160 }}
          allowClear
          onClear={() => { setIp(''); setPage(1); }}
        />
        <RangePicker
          onChange={(dates) => setDateRange(dates as [Dayjs | null, Dayjs | null] | null)}
        />
        <Button type="primary" icon={<SearchOutlined />} onClick={handleFilter}>
          {t('activityLogs.filter')}
        </Button>
      </Space>

      {error && (
        <div style={{ padding: 24 }}>
          <Result status="error" title={t('activityLogs.failed')} subTitle={error.message} />
        </div>
      )}

      <Table
        rowKey="id"
        columns={columns}
        dataSource={logs}
        loading={isLoading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (tCount) => t('activityLogs.total_0', { count: tCount }),
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
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
            <Descriptions.Item label={t('activityLogs.extra')}>{detailLog.extra ? JSON.stringify(detailLog.extra, null, 2) : '-'}</Descriptions.Item>
            <Descriptions.Item label={t('activityLogs.time')}>
              {new Date(detailLog.createdAt).toLocaleString()}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}
