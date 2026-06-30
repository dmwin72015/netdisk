import { useEffect, useState, useMemo } from 'react';
import {
  Table,
  Input,
  Select,
  Button,
  Space,
  message,
  Popconfirm,
  Tag,
} from 'antd';
import { SearchOutlined, DeleteOutlined } from '@ant-design/icons';
import { adminListFiles, adminDeleteFiles, type AdminFile } from '../api/admin';
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
  const [data, setData] = useState<AdminFile[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [userFilter, setUserFilter] = useState<string | undefined>(undefined);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await adminListFiles(pageSize, (page - 1) * pageSize);
      let items = res.items;
      if (userFilter) {
        items = items.filter((f) => f.userId === userFilter);
      }
      if (search.trim()) {
        const q = search.trim().toLowerCase();
        items = items.filter((f) => f.fileName.toLowerCase().includes(q));
      }
      setData(items);
      setTotal(res.total);
    } catch {
      message.error('Failed to load files');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [page, pageSize]);

  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) return;
    try {
      await adminDeleteFiles(selectedRowKeys.map(String));
      message.success(`${selectedRowKeys.length} files deleted`);
      setSelectedRowKeys([]);
      loadData();
    } catch {
      message.error('Failed to delete files');
    }
  };

  const columns: ColumnsType<AdminFile> = [
    { title: 'Filename', dataIndex: 'fileName', key: 'fileName', ellipsis: true },
    { title: 'User', dataIndex: 'username', key: 'username', width: 150 },
    {
      title: 'Type',
      dataIndex: 'fileCategory',
      key: 'fileCategory',
      width: 120,
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
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
  };

  const userOptions = useMemo(() => {
    const map = new Map<string, string>();
    data.forEach((f) => {
      if (!map.has(f.userId)) map.set(f.userId, f.username);
    });
    return Array.from(map.entries()).map(([id, name]) => ({ label: name, value: id }));
  }, [data]);

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
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onPressEnter={() => { setPage(1); loadData(); }}
            style={{ width: 220 }}
            allowClear
          />
          <Select
            placeholder="Filter by user"
            allowClear
            value={userFilter}
            onChange={(val) => { setUserFilter(val); setPage(1); }}
            style={{ width: 180 }}
            options={userOptions}
          />
          <Button onClick={() => { setPage(1); loadData(); }}>Search</Button>
        </Space>
        {selectedRowKeys.length > 0 && (
          <Popconfirm
            title={`Delete ${selectedRowKeys.length} files?`}
            onConfirm={handleBatchDelete}
            okText="Yes"
            cancelText="No"
          >
            <Button danger icon={<DeleteOutlined />}>
              Batch Delete ({selectedRowKeys.length})
            </Button>
          </Popconfirm>
        )}
      </div>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        rowSelection={rowSelection}
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
