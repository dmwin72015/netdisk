import { useRef } from 'react';
import { Space, Tag, Button, Popconfirm, message } from 'antd';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { UndoOutlined, DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router';
import { PageContainer } from '@/components/PageContainer';
import CopyCell from '@/components/CopyCell';
import { useDeleteFile, useRestoreFile, fetchFiles } from '@/api/files';
import type { AdminFile } from '@/api/files';
import { formatBytes, formatDate } from '@/utils/format';

const CATEGORY_OPTIONS = [
  { label: '文档', value: 'document' },
  { label: '图片', value: 'image' },
  { label: '视频', value: 'video' },
  { label: '音频', value: 'audio' },
  { label: '压缩包', value: 'archive' },
  { label: '其他', value: 'other' },
];
const STATUS_OPTIONS = [
  { label: '正常', value: 'active' },
  { label: '已删除', value: 'trashed' },
];

export default function FilesPage() {
  const actionRef = useRef<ActionType>(null);
  const deleteFileMut = useDeleteFile();
  const restoreFileMut = useRestoreFile();

  const columns: ProColumns<AdminFile>[] = [
    { title: 'ID', dataIndex: 'id', width: 60, hideInSearch: true },
    { title: 'Slug', dataIndex: 'slug', width: 220, hideInSearch: true, render: (_, r) => <CopyCell value={r.slug} /> },
    { title: '文件名', dataIndex: 'fileName', ellipsis: true },
    {
      title: '所属用户', dataIndex: 'username', width: 200, hideInSearch: true,
      render: (_, r) => <Link to={`/users/${r.userId}`}>{r.username}</Link>,
    },
    {
      title: '类型', dataIndex: 'fileCategory', valueType: 'select', width: 100,
      fieldProps: { options: CATEGORY_OPTIONS },
      render: (_, r) => {
        const opt = CATEGORY_OPTIONS.find(o => o.value === r.fileCategory);
        return <Tag color="blue">{opt?.label || '其他'}</Tag>;
      },
    },
    {
      title: '大小', dataIndex: 'fileSize', width: 160, hideInSearch: true, sorter: true,
      render: (_, r) => r.isDir ? <span className="text-[#999]">-</span> : formatBytes(r.fileSize),
    },
    {
      title: '状态', dataIndex: 'isTrashed', valueType: 'select', width: 140,
      fieldProps: { options: STATUS_OPTIONS },
      render: (_, r) => (
        <Space>
          {r.isTrashed && <Tag color="red">已删除</Tag>}
          {r.isStarred && <Tag color="gold">收藏</Tag>}
          {!r.isTrashed && !r.isStarred && <Tag>正常</Tag>}
        </Space>
      ),
    },
    {
      title: '上传时间', dataIndex: 'createdAt', width: 160, hideInSearch: true, sorter: true,
      render: (_, record) => formatDate(record.createdAt),
    },
    {
      title: '操作', width: 180, hideInSearch: true, fixed: 'right',
      render: (_, r) => (
        <Space>
          {r.isTrashed && (
            <Popconfirm title="恢复文件" description="确定要恢复此文件吗？" okText="确定" cancelText="取消"
              onConfirm={() => restoreFileMut.mutate(r.id, { onSuccess: () => { message.success('已恢复'); actionRef.current?.reload(); }, onError: () => message.error('恢复失败') })}
            >
              <Button type="link" size="small" icon={<UndoOutlined />}>恢复</Button>
            </Popconfirm>
          )}
          <Popconfirm title="永久删除" description="此操作不可撤销" okText="确定" cancelText="取消"
            onConfirm={() => deleteFileMut.mutate(r.id, { onSuccess: () => { message.success('已删除'); actionRef.current?.reload(); }, onError: () => message.error('删除失败') })}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title="用户文件">
      <ProTable<AdminFile>
        headerTitle="文件列表"
        rowKey="id"
        actionRef={actionRef}
        columns={columns}
        request={async (params, sorter) => {
          const sortField = Object.keys(sorter)[0];
          const sortOrder = sortField ? (sorter as Record<string, string>)[sortField] === 'ascend' ? 'asc' : 'desc' : undefined;
          const res = await fetchFiles({
            limit: params.pageSize!,
            offset: (params.current! - 1) * params.pageSize!,
            search: (params.fileName as string) || undefined,
            fileCategory: (params.fileCategory as string) || undefined,
            isTrashed: params.isTrashed === 'trashed' ? true : params.isTrashed === 'active' ? false : undefined,
            sortBy: sortField as string | undefined,
            sortOrder,
          });
          return { data: res.items, success: true, total: res.total };
        }}
        search={{ labelWidth: 'auto' }}
        pagination={{ showSizeChanger: true, showTotal: (total) => `共 ${total} 条` }}
        size="small"
        options={false}
        scroll={{ x: 'max-content' }}
      />
    </PageContainer>
  );
}
