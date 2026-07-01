import { useRef } from 'react';
import { Button, Space, Tag, Popconfirm, message } from 'antd';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { UndoOutlined, DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useDeleteFile, useRestoreFile } from '../../../api/admin.hooks';
import { fetchFiles } from '../../../api/admin';
import type { AdminFile } from '../../../api/admin';

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(epoch: number): string {
  return new Date(epoch * 1000).toLocaleString();
}

const CATEGORY_OPTIONS = [
  { label: 'document', value: 'document' },
  { label: 'image', value: 'image' },
  { label: 'video', value: 'video' },
  { label: 'audio', value: 'audio' },
  { label: 'archive', value: 'archive' },
  { label: 'other', value: 'other' },
];

const STATUS_OPTIONS = [
  { label: 'active', value: 'active' },
  { label: 'trashed', value: 'trashed' },
];

export default function Files() {
  const { t } = useTranslation();
  const actionRef = useRef<ActionType>();
  const deleteFileMut = useDeleteFile();
  const restoreFileMut = useRestoreFile();

  const columns: ProColumns<AdminFile>[] = [
    { title: t('files.filename'), dataIndex: 'fileName', ellipsis: true },
    {
      title: t('files.owner'),
      dataIndex: 'username',
      render: (_, r) => <Link to={`/admin/users/${r.userId}`}>{r.username}</Link>,
      width: 150,
      hideInSearch: true,
    },
    {
      title: t('files.type'),
      dataIndex: 'fileCategory',
      valueType: 'select',
      fieldProps: { options: CATEGORY_OPTIONS },
      render: (v) => <Tag color="blue">{v || t('files.other')}</Tag>,
      width: 100,
    },
    {
      title: t('files.size'),
      dataIndex: 'fileSize',
      render: (v: number) => formatBytes(v),
      width: 100,
      hideInSearch: true,
      sorter: true,
    },
    {
      title: t('files.status'),
      dataIndex: 'isTrashed',
      valueType: 'select',
      fieldProps: { options: STATUS_OPTIONS },
      render: (_, r) => (
        <Space>
          {r.isTrashed && <Tag color="red">{t('files.deleted')}</Tag>}
          {r.isStarred && <Tag color="gold">{t('files.starred')}</Tag>}
          {!r.isTrashed && !r.isStarred && <Tag>{t('files.normal')}</Tag>}
        </Space>
      ),
      width: 140,
    },
    {
      title: t('files.uploaded'),
      dataIndex: 'createdAt',
      render: (v: number) => formatDate(v),
      width: 160,
      hideInSearch: true,
      sorter: true,
    },
    {
      title: t('files.actions'),
      width: 180,
      hideInSearch: true,
      render: (_, r) => (
        <Space>
          {r.isTrashed && (
            <Popconfirm
              title={t('files.restore')}
              description={t('files.restoreConfirm')}
              onConfirm={() =>
                restoreFileMut.mutate(r.id, {
                  onSuccess: () => {
                    message.success(t('files.restoreSuccess'));
                    actionRef.current?.reload();
                  },
                  onError: () => message.error(t('files.restoreFailed')),
                })
              }
              okText={t('common.yes')}
              cancelText={t('common.no')}
            >
              <Button type="link" size="small" icon={<UndoOutlined />}>
                {t('files.restore')}
              </Button>
            </Popconfirm>
          )}
          <Popconfirm
            title={t('files.permanentDelete')}
            description={t('files.noUndo')}
            onConfirm={() =>
              deleteFileMut.mutate(r.id, {
                onSuccess: () => {
                  message.success(t('files.deleteSuccess'));
                  actionRef.current?.reload();
                },
                onError: () => message.error(t('files.deleteFailed')),
              })
            }
            okText={t('common.yes')}
            cancelText={t('common.no')}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <ProTable<AdminFile>
      rowKey="id"
      actionRef={actionRef}
      columns={columns}
      request={async (params, sorter) => {
        const sortField = Object.keys(sorter)[0];
        const sortOrder = sortField
          ? (sorter as Record<string, string>)[sortField] === 'ascend'
            ? 'asc'
            : 'desc'
          : undefined;
        const res = await fetchFiles({
          limit: params.pageSize!,
          offset: (params.current! - 1) * params.pageSize!,
          search: (params.fileName as string) || undefined,
          fileCategory: (params.fileCategory as string) || undefined,
          isTrashed:
            params.isTrashed === 'trashed'
              ? true
              : params.isTrashed === 'active'
                ? false
                : undefined,
          sortBy: sortField as string | undefined,
          sortOrder,
        });
        return { data: res.items, success: true, total: res.total };
      }}
      search={{ labelWidth: 'auto' }}
      pagination={{
        showSizeChanger: true,
        showTotal: (total) => t('files.total_0', { count: total }),
      }}
      size="small"
      options={false}
    />
  );
}