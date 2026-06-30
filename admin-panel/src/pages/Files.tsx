import { useState } from 'react';
import {
  Table,
  Input,
  Select,
  Button,
  Space,
  message,
  Popconfirm,
  Tag,
  Result,
} from 'antd';
import { SearchOutlined, DeleteOutlined, UndoOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useFiles, useDeleteFile, useRestoreFile } from '../api/admin.hooks';
import type { AdminFile } from '../api/admin';
import type { ColumnsType } from 'antd/es/table';

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

export default function Files() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string | undefined>(undefined);
  const [trashedFilter, setTrashedFilter] = useState<boolean | undefined>(undefined);
  const [sortValue, setSortValue] = useState('createdAt-desc');

  const [sortBy, sortOrder] = sortValue.split('-') as [string, 'asc' | 'desc'];

  const { data, isLoading, error } = useFiles({
    limit: pageSize,
    offset: (page - 1) * pageSize,
    ...(searchQuery.trim() ? { search: searchQuery.trim() } : {}),
    ...(categoryFilter ? { fileCategory: categoryFilter } : {}),
    ...(trashedFilter !== undefined ? { isTrashed: trashedFilter } : {}),
    sortBy,
    sortOrder,
  });

  const CATEGORY_OPTIONS = [
    { label: t('files.allCategories'), value: '' },
    { label: t('files.document'), value: 'document' },
    { label: t('files.image'), value: 'image' },
    { label: t('files.video'), value: 'video' },
    { label: t('files.audio'), value: 'audio' },
    { label: t('files.archive'), value: 'archive' },
    { label: t('files.other'), value: 'other' },
  ];

  const TRASHED_OPTIONS = [
    { label: t('files.allStatus'), value: '' },
    { label: t('files.active'), value: 'active' },
    { label: t('files.trashed'), value: 'trashed' },
  ];

  const SORT_OPTIONS = [
    { label: t('files.sortNewest'), value: 'createdAt-desc' },
    { label: t('files.sortOldest'), value: 'createdAt-asc' },
    { label: t('files.sortNameAsc'), value: 'fileName-asc' },
    { label: t('files.sortNameDesc'), value: 'fileName-desc' },
    { label: t('files.sortSizeDesc'), value: 'fileSize-desc' },
    { label: t('files.sortSizeAsc'), value: 'fileSize-asc' },
  ];

  const deleteFileMut = useDeleteFile();
  const restoreFileMut = useRestoreFile();

  const files = data?.items ?? [];
  const total = data?.total ?? 0;

  const handleDelete = async (id: string) => {
    try {
      await deleteFileMut.mutateAsync(id);
      message.success(t('files.deleteSuccess'));
    } catch {
      message.error(t('files.deleteFailed'));
    }
  };

  const handleRestore = async (id: string) => {
    try {
      await restoreFileMut.mutateAsync(id);
      message.success(t('files.restoreSuccess'));
    } catch {
      message.error(t('files.restoreFailed'));
    }
  };

  const handleTrashedFilter = (val: string) => {
    if (val === '' || val === undefined) {
      setTrashedFilter(undefined);
    } else if (val === 'active') {
      setTrashedFilter(false);
    } else {
      setTrashedFilter(true);
    }
    setPage(1);
  };

  const columns: ColumnsType<AdminFile> = [
    { title: t('files.filename'), dataIndex: 'fileName', key: 'fileName', ellipsis: true },
    {
      title: t('files.owner'),
      dataIndex: 'username',
      key: 'username',
      width: 150,
      render: (text: string, record: AdminFile) => (
        <a onClick={() => navigate(`/admin/users/${record.userId}`)} style={{ cursor: 'pointer' }}>
          {text}
        </a>
      ),
    },
    {
      title: t('files.type'),
      dataIndex: 'fileCategory',
      key: 'fileCategory',
      width: 100,
      render: (v: string) => <Tag color="blue">{v || t('files.other')}</Tag>,
    },
    {
      title: t('files.size'),
      dataIndex: 'fileSize',
      key: 'fileSize',
      width: 100,
      render: (v: number) => formatBytes(v),
    },
    {
      title: t('files.status'),
      key: 'status',
      width: 140,
      render: (_: unknown, record: AdminFile) => (
        <Space>
          {record.isTrashed && <Tag color="red">{t('files.deleted')}</Tag>}
          {record.isStarred && <Tag color="gold">{t('files.starred')}</Tag>}
          {!record.isTrashed && !record.isStarred && <Tag>{t('files.normal')}</Tag>}
        </Space>
      ),
    },
    {
      title: t('files.uploaded'),
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 160,
      render: (v: number) => formatDate(v),
    },
    {
      title: t('files.actions'),
      key: 'actions',
      width: 180,
      render: (_: unknown, record: AdminFile) => (
        <Space>
          {record.isTrashed && (
            <Popconfirm
              title={t('files.restore')}
              description={t('files.restoreConfirm')}
              onConfirm={() => handleRestore(record.id)}
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
            onConfirm={() => handleDelete(record.id)}
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
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          marginBottom: 16,
          flexWrap: 'wrap',
          gap: 8,
        }}
      >
        <Space wrap>
          <Input
            placeholder={t('files.search')}
            prefix={<SearchOutlined />}
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onPressEnter={() => {
              setSearchQuery(searchInput);
              setPage(1);
            }}
            style={{ width: 220 }}
            allowClear
            onClear={() => {
              setSearchInput('');
              setSearchQuery('');
              setPage(1);
            }}
          />
          <Select
            placeholder={t('files.category')}
            allowClear
            value={categoryFilter || ''}
            onChange={(val) => {
              setCategoryFilter(val || undefined);
              setPage(1);
            }}
            style={{ width: 130 }}
            options={CATEGORY_OPTIONS}
          />
          <Select
            placeholder={t('files.statusFilter')}
            allowClear
            value={trashedFilter === undefined ? '' : trashedFilter ? 'trashed' : 'active'}
            onChange={handleTrashedFilter}
            style={{ width: 130 }}
            options={TRASHED_OPTIONS}
          />
          <Select
            value={sortValue}
            onChange={(val: string) => {
              setSortValue(val);
              setPage(1);
            }}
            style={{ width: 150 }}
            options={SORT_OPTIONS}
          />
          <Button onClick={() => { setSearchQuery(searchInput); setPage(1); }}>{t('files.searchButton')}</Button>
        </Space>
      </div>
      {error && (
        <div style={{ padding: 24 }}>
          <Result status="error" title={t('files.failed')} subTitle={error.message} />
        </div>
      )}
      <Table
        rowKey="id"
        columns={columns}
        dataSource={files}
        loading={isLoading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (tCount) => t('files.total_0', { count: tCount }),
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
        size="small"
      />
    </div>
  );
}