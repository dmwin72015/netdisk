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

const CATEGORY_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Document', value: 'document' },
  { label: 'Image', value: 'image' },
  { label: 'Video', value: 'video' },
  { label: 'Audio', value: 'audio' },
  { label: 'Archive', value: 'archive' },
  { label: 'Other', value: 'other' },
];

const TRASHED_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Active', value: 'active' },
  { label: 'Trashed', value: 'trashed' },
];

const SORT_OPTIONS = [
  { label: 'Newest first', value: 'createdAt-desc' },
  { label: 'Oldest first', value: 'createdAt-asc' },
  { label: 'Name A-Z', value: 'fileName-asc' },
  { label: 'Name Z-A', value: 'fileName-desc' },
  { label: 'Largest first', value: 'fileSize-desc' },
  { label: 'Smallest first', value: 'fileSize-asc' },
];

export default function Files() {
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

  const deleteFileMut = useDeleteFile();
  const restoreFileMut = useRestoreFile();

  const files = data?.items ?? [];
  const total = data?.total ?? 0;

  const handleDelete = async (id: string) => {
    try {
      await deleteFileMut.mutateAsync(id);
      message.success('File permanently deleted');
    } catch {
      message.error('Failed to delete file');
    }
  };

  const handleRestore = async (id: string) => {
    try {
      await restoreFileMut.mutateAsync(id);
      message.success('File restored');
    } catch {
      message.error('Failed to restore file');
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
    { title: 'Filename', dataIndex: 'fileName', key: 'fileName', ellipsis: true },
    {
      title: 'Owner',
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
      title: 'Type',
      dataIndex: 'fileCategory',
      key: 'fileCategory',
      width: 100,
      render: (v: string) => <Tag color="blue">{v || 'file'}</Tag>,
    },
    {
      title: 'Size',
      dataIndex: 'fileSize',
      key: 'fileSize',
      width: 100,
      render: (v: number) => formatBytes(v),
    },
    {
      title: 'Status',
      key: 'status',
      width: 140,
      render: (_: unknown, record: AdminFile) => (
        <Space>
          {record.isTrashed && <Tag color="red">Trashed</Tag>}
          {record.isStarred && <Tag color="gold">Starred</Tag>}
          {!record.isTrashed && !record.isStarred && <Tag>Normal</Tag>}
        </Space>
      ),
    },
    {
      title: 'Uploaded',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 160,
      render: (v: number) => formatDate(v),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 180,
      render: (_: unknown, record: AdminFile) => (
        <Space>
          {record.isTrashed && (
            <Popconfirm
              title="Restore file"
              description="Restore this file from trash?"
              onConfirm={() => handleRestore(record.id)}
              okText="Yes"
              cancelText="No"
            >
              <Button type="link" size="small" icon={<UndoOutlined />}>
                Restore
              </Button>
            </Popconfirm>
          )}
          <Popconfirm
            title="Permanently delete"
            description="This action cannot be undone."
            onConfirm={() => handleDelete(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              Delete
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
            placeholder="Search files..."
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
            placeholder="Category"
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
            placeholder="Status"
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
          <Button onClick={() => { setSearchQuery(searchInput); setPage(1); }}>Search</Button>
        </Space>
      </div>
      {error && (
        <div style={{ padding: 24 }}>
          <Result status="error" title="Failed to load files" subTitle={error.message} />
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
          showTotal: (t) => `Total ${t}`,
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