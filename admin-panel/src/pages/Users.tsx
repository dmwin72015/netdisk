import { useEffect, useState } from 'react';
import {
  Table,
  Input,
  Select,
  Button,
  Space,
  Modal,
  Form,
  InputNumber,
  message,
  Popconfirm,
  Tooltip,
} from 'antd';
import { SearchOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import {
  adminListUsers,
  adminUpdateRole,
  adminUpdateStorageBase,
  adminDeleteUser,
  adminDeleteUsers,
  type AdminUser,
} from '../api/admin';
import type { ColumnsType } from 'antd/es/table';

const ROLES = ['admin', 'user', 'moderator'];

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(epoch: number): string {
  return new Date(epoch * 1000).toLocaleDateString();
}

export default function Users() {
  const navigate = useNavigate();
  const [data, setData] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [storageModal, setStorageModal] = useState<{ open: boolean; user: AdminUser | null }>({
    open: false,
    user: null,
  });
  const [storageForm] = Form.useForm();

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await adminListUsers(pageSize, (page - 1) * pageSize);
      let items = res.items;
      if (roleFilter) {
        items = items.filter((u) => u.role === roleFilter);
      }
      if (search.trim()) {
        const q = search.trim().toLowerCase();
        items = items.filter(
          (u) =>
            u.username.toLowerCase().includes(q) ||
            u.email.toLowerCase().includes(q)
        );
      }
      setData(items);
      setTotal(res.total);
    } catch {
      message.error('Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [page, pageSize]);

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      const updated = await adminUpdateRole(userId, role);
      setData((prev) =>
        prev.map((u) => (u.id === userId ? updated : u))
      );
      message.success('Role updated');
    } catch {
      message.error('Failed to update role');
      loadData();
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await adminDeleteUser(id);
      message.success('User deleted');
      setData((prev) => prev.filter((u) => u.id !== id));
      setTotal((t) => t - 1);
    } catch {
      message.error('Failed to delete user');
    }
  };

  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) return;
    try {
      await adminDeleteUsers(selectedRowKeys.map(String));
      message.success(`${selectedRowKeys.length} users deleted`);
      setSelectedRowKeys([]);
      loadData();
    } catch {
      message.error('Failed to batch delete');
    }
  };

  const handleStorageSave = async (values: { baseBytes: number }) => {
    if (!storageModal.user) return;
    try {
      const updated = await adminUpdateStorageBase(storageModal.user.id, values.baseBytes);
      setData((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
      message.success('Storage base updated');
      setStorageModal({ open: false, user: null });
    } catch {
      message.error('Failed to update storage');
    }
  };

  const columns: ColumnsType<AdminUser> = [
    {
      title: 'Username',
      dataIndex: 'username',
      key: 'username',
      render: (text: string, record: AdminUser) => (
        <a onClick={() => navigate(`/users/${record.id}`)} style={{ cursor: 'pointer' }}>{text}</a>
      ),
    },
    { title: 'Email', dataIndex: 'email', key: 'email' },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      width: 130,
      render: (role: string, record: AdminUser) => (
        <Select
          value={role}
          size="small"
          style={{ width: 120 }}
          onChange={(val) => handleRoleChange(record.id, val)}
          options={ROLES.map((r) => ({ label: r, value: r }))}
        />
      ),
    },
    {
      title: 'Storage Used',
      key: 'storage',
      width: 130,
      render: (_: unknown, record: AdminUser) => (
        <Tooltip title={`Total: ${formatBytes(record.totalBytes)}`}>
          {formatBytes(record.usedBytes)}
        </Tooltip>
      ),
    },
    {
      title: 'Registered',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 120,
      render: (v: number) => formatDate(v),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_: unknown, record: AdminUser) => (
        <Space>
          <Tooltip title="View">
            <EyeOutlined
              style={{ cursor: 'pointer', color: '#1890ff' }}
              onClick={() => navigate(`/users/${record.id}`)}
            />
          </Tooltip>
          <Popconfirm
            title="Delete user"
            description="Are you sure?"
            onConfirm={() => handleDelete(record.id)}
            okText="Yes"
            cancelText="No"
          >
            <DeleteOutlined style={{ cursor: 'pointer', color: '#ff4d4f' }} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => setSelectedRowKeys(keys),
  };

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
            placeholder="Search users..."
            prefix={<SearchOutlined />}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onPressEnter={() => {
              setPage(1);
              loadData();
            }}
            style={{ width: 220 }}
            allowClear
          />
          <Select
            placeholder="Filter by role"
            allowClear
            value={roleFilter}
            onChange={(val) => {
              setRoleFilter(val);
              setPage(1);
            }}
            style={{ width: 150 }}
            options={ROLES.map((r) => ({ label: r, value: r }))}
          />
          <Button onClick={() => { setPage(1); loadData(); }}>Search</Button>
        </Space>
        {selectedRowKeys.length > 0 && (
          <Popconfirm
            title={`Delete ${selectedRowKeys.length} users?`}
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
      <Modal
        title={`Edit Storage Base - ${storageModal.user?.username}`}
        open={storageModal.open}
        onCancel={() => setStorageModal({ open: false, user: null })}
        footer={null}
      >
        <Form form={storageForm} layout="vertical" onFinish={handleStorageSave}>
          <Form.Item
            name="baseBytes"
            label="Base Bytes"
            initialValue={storageModal.user?.baseBytes}
            rules={[{ required: true, message: 'Required' }]}
          >
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              Save
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
