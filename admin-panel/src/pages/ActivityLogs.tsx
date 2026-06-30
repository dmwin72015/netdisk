import { useState } from 'react';
import { Table, Input, Select, Button, Space, Modal, Descriptions, DatePicker, Result } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useActivityLogs, useActivityLogActions } from '../api/admin.hooks';
import type { ActivityLogParams, AdminActivityLog } from '../api/admin';
import type { ColumnsType } from 'antd/es/table';
import type { Dayjs } from 'dayjs';

const { RangePicker } = DatePicker;

export default function ActivityLogs() {
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
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: 'User',
      dataIndex: 'username',
      render: (_, r) => <Link to={`/admin/users/${r.userId}`}>{r.username}</Link>,
    },
    { title: 'Action', dataIndex: 'actionLabel' },
    {
      title: 'Resource',
      render: (_, r) => `${r.resourceType}: ${r.resourceName}`,
    },
    { title: 'IP', dataIndex: 'ip' },
    { title: 'OS', dataIndex: 'os' },
    { title: 'Browser', dataIndex: 'browser' },
    {
      title: 'Time',
      dataIndex: 'createdAt',
      width: 180,
      render: (v: string) => new Date(v).toLocaleString(),
    },
    {
      title: '',
      width: 80,
      render: (_, r) => (
        <Button size="small" onClick={() => setDetailLog(r)}>
          Detail
        </Button>
      ),
    },
  ];

  const handleFilter = () => {
    setPage(1);
  };

  return (
    <div>
      <h2 style={{ marginBottom: 16 }}>Activity Logs</h2>

      <Space wrap style={{ marginBottom: 16 }}>
        <Input
          placeholder="User ID"
          type="number"
          value={userId ?? ''}
          onChange={(e) => setUserId(e.target.value || undefined)}
          style={{ width: 120 }}
          allowClear
          onClear={() => { setUserId(undefined); setPage(1); }}
        />
        <Select
          placeholder="Action"
          allowClear
          value={action}
          onChange={(val) => { setAction(val); setPage(1); }}
          style={{ width: 180 }}
          options={(actions ?? []).map((a) => ({ label: a.label, value: a.action }))}
        />
        <Input
          placeholder="IP"
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
          Filter
        </Button>
      </Space>

      {error && (
        <div style={{ padding: 24 }}>
          <Result status="error" title="Failed to load activity logs" subTitle={error.message} />
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
          showTotal: (t) => `Total ${t}`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
        size="small"
      />

      <Modal
        title="Activity Log Detail"
        open={!!detailLog}
        onCancel={() => setDetailLog(null)}
        footer={null}
        width={600}
      >
        {detailLog && (
          <Descriptions column={1} size="small">
            <Descriptions.Item label="ID">{detailLog.id}</Descriptions.Item>
            <Descriptions.Item label="User">
              {detailLog.username} ({detailLog.userId})
            </Descriptions.Item>
            <Descriptions.Item label="Action">{detailLog.actionLabel}</Descriptions.Item>
            <Descriptions.Item label="Resource">
              {detailLog.resourceType}: {detailLog.resourceName}
            </Descriptions.Item>
            <Descriptions.Item label="IP">{detailLog.ip || '-'}</Descriptions.Item>
            <Descriptions.Item label="IP Region">{detailLog.ipRegion || '-'}</Descriptions.Item>
            <Descriptions.Item label="User Agent">{detailLog.userAgent || '-'}</Descriptions.Item>
            <Descriptions.Item label="OS">{detailLog.os || '-'}</Descriptions.Item>
            <Descriptions.Item label="Browser">{detailLog.browser || '-'}</Descriptions.Item>
            <Descriptions.Item label="Extra">{detailLog.extra ? JSON.stringify(detailLog.extra, null, 2) : '-'}</Descriptions.Item>
            <Descriptions.Item label="Time">
              {new Date(detailLog.createdAt).toLocaleString()}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}
