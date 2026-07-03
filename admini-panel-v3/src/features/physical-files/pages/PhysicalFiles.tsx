import { useRef } from 'react';
import { Button, Space, Tag } from 'antd';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router';
import dayjs from 'dayjs';
import { PageContainer } from '@/components/PageContainer';
import CopyCell from '@/components/CopyCell';
import { fetchPhysicalFiles } from '@/api/admin';
import type { AdminPhysicalFile } from '@/api/admin';
import { formatBytes, formatDate } from '@/utils/format';

const STATUS_OPTIONS = [
  { label: '已完成', value: 'completed' },
  { label: '处理中', value: 'processing' },
  { label: '失败', value: 'failed' },
];
const STATUS_COLORS: Record<string, string> = { completed: 'green', processing: 'blue', failed: 'red' };

export default function PhysicalFilesPage() {
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);

  const columns: ProColumns<AdminPhysicalFile>[] = [
    { title: 'ID', dataIndex: 'id', width: 60, hideInSearch: true },
    { title: 'Slug', dataIndex: 'slug', width: 140, render: (_, r) => <CopyCell value={r.slug} /> },
    { title: '文件哈希', dataIndex: 'fileHash', width: 240, render: (_, r) => <CopyCell value={r.fileHash}>{r.fileHash.slice(0, 16)}...</CopyCell> },
    { title: '文件大小', dataIndex: 'fileSize', width: 100, hideInSearch: true, render: (_, r) => formatBytes(r.fileSize) },
    { title: 'MIME 类型', dataIndex: 'mimeType', width: 120 },
    { title: '存储路径', dataIndex: 'storagePath', width: 200, ellipsis: true, hideInSearch: true },
    {
      title: '状态', dataIndex: 'status', width: 100, valueType: 'select',
      fieldProps: { options: STATUS_OPTIONS },
      render: (_, r) => <Tag color={STATUS_COLORS[r.status] || 'default'}>{STATUS_OPTIONS.find(o => o.value === r.status)?.label ?? r.status}</Tag>,
    },
    {
      title: '引用次数', key: 'referenceCount', width: 110, hideInSearch: true,
      render: (_, r) => <span>{r.userFileCount + r.mediaItemCount}</span>,
    },
    {
      title: '创建时间', dataIndex: 'createdAt', width: 160, valueType: 'dateTimeRange', hideInSearch: true,
      render: (_, r) => formatDate(r.createdAt),
    },
    {
      title: '操作', width: 120, fixed: 'right', hideInSearch: true,
      render: (_, r) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/files/physical/${r.id}`)}>详情</Button>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title="物理文件">
      <ProTable<AdminPhysicalFile>
        headerTitle="物理文件列表"
        rowKey="id"
        actionRef={actionRef}
        columns={columns}
        request={async (params) => {
          const fileSize = params.fileSize as [number, number] | undefined;
          const createdAt = params.createdAt as [dayjs.Dayjs, dayjs.Dayjs] | undefined;
          const res = await fetchPhysicalFiles({
            limit: params.pageSize!,
            offset: (params.current! - 1) * params.pageSize!,
            search: (params.slug as string) || undefined,
            status: (params.status as string) || undefined,
            hash_filter: (params.fileHash as string) || undefined,
            mime_filter: (params.mimeType as string) || undefined,
            min_size: fileSize?.[0],
            max_size: fileSize?.[1],
            created_from: createdAt?.[0]?.toISOString(),
            created_to: createdAt?.[1]?.toISOString(),
          });
          return { data: res.items, success: true, total: res.total };
        }}
        search={{ labelWidth: 'auto' }}
        scroll={{ x: 1300 }}
        pagination={{ showSizeChanger: true, showTotal: (total) => `共 ${total} 条` }}
        size="small"
        options={false}
      />
    </PageContainer>
  );
}
