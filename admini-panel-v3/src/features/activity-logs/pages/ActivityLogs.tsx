import { useState } from 'react';
import { Button, Modal } from 'antd';
import { ProDescriptions, ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { EyeOutlined } from '@ant-design/icons';
import { Link } from 'react-router';
import { PageContainer } from '@/components/PageContainer';
import { useActivityLogActions } from '@/api/admin.hooks';
import { fetchActivityLogs } from '@/api/admin';
import type { AdminActivityLog } from '@/api/admin';
import { formatISODate } from '@/utils/format';

export default function ActivityLogsPage() {
  const [detailLog, setDetailLog] = useState<AdminActivityLog | null>(null);
  const { data: actions } = useActivityLogActions();

  const columns: ProColumns<AdminActivityLog>[] = [
    { title: 'ID', dataIndex: 'id', width: 60, hideInSearch: true },
    {
      title: '用户', dataIndex: 'username', width: 120, hideInSearch: true,
      render: (_, r) => <Link to={`/users/${r.userId}`}>{r.username}</Link>,
    },
    {
      title: '操作', dataIndex: 'action', width: 140, valueType: 'select',
      fieldProps: { options: (actions ?? []).map(a => ({ label: a.label, value: a.action })) },
      render: (_, r) => r.actionLabel,
    },
    {
      title: '资源', width: 180, hideInSearch: true,
      render: (_, r) => `${r.resourceType}: ${r.resourceName}`,
    },
    { title: 'IP 地址', dataIndex: 'ip', width: 130 },
    { title: '系统', dataIndex: 'os', width: 100, hideInSearch: true },
    { title: '浏览器', dataIndex: 'browser', width: 140, hideInSearch: true, ellipsis: true },
    {
      title: '时间', dataIndex: 'createdAt', width: 170, hideInSearch: true, sorter: true, defaultSortOrder: 'descend',
      render: (_, record) => formatISODate(record.createdAt),
    },
    {
      title: '', width: 80, hideInSearch: true, fixed: 'right',
      render: (_, r) => <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => setDetailLog(r)}>详情</Button>,
    },
  ];

  return (
    <PageContainer title="操作日志">
      <ProTable<AdminActivityLog>
        headerTitle="操作日志"
        rowKey="id"
        columns={columns}
        request={async (params, sorter) => {
          const sortField = Object.keys(sorter)[0];
          const sortOrder = sortField ? (sorter as Record<string, string>)[sortField] === 'ascend' ? 'asc' : 'desc' : undefined;
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
        search={{ labelWidth: 'auto', defaultCollapsed: false }}
        pagination={{ showSizeChanger: true, showTotal: (total) => `共 ${total} 条` }}
        size="small"
        scroll={{ x: 'max-content' }}
      />

      <Modal title="日志详情" open={!!detailLog} onCancel={() => setDetailLog(null)} footer={null} width={600}>
        {detailLog && (
          <ProDescriptions column={1} size="small">
            <ProDescriptions.Item label="ID">{detailLog.id}</ProDescriptions.Item>
            <ProDescriptions.Item label="用户">{detailLog.username} ({detailLog.userId})</ProDescriptions.Item>
            <ProDescriptions.Item label="操作">{detailLog.actionLabel}</ProDescriptions.Item>
            <ProDescriptions.Item label="资源">{detailLog.resourceType}: {detailLog.resourceName}</ProDescriptions.Item>
            <ProDescriptions.Item label="IP 地址">{detailLog.ip || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label="IP 归属">{detailLog.ipRegion || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label="User-Agent">{detailLog.userAgent || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label="系统">{detailLog.os || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label="浏览器">{detailLog.browser || '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label="附加数据">{detailLog.extra ? JSON.stringify(detailLog.extra, null, 2) : '-'}</ProDescriptions.Item>
            <ProDescriptions.Item label="时间">{formatISODate(detailLog.createdAt)}</ProDescriptions.Item>
          </ProDescriptions>
        )}
      </Modal>
    </PageContainer>
  );
}
