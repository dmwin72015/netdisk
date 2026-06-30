import { useState } from 'react';
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
import {
  SearchOutlined,
  DeleteOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import {
  useUsers,
  useCreateUser,
  useUpdateUserRole,
  useUpdateStorageBase,
  useDeleteUser,
} from '../api/admin.hooks';
import type { AdminUser, CreateUserInput } from '../api/admin';
import type { ColumnsType } from 'antd/es/table';

const ROLES = ['admin', 'user'];

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

type SortValue = 'newest' | 'oldest' | 'name-asc' | 'name-desc';

export default function Users() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [searchInput, setSearchInput] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState<string | undefined>(undefined);
  const [sortBy, setSortBy] = useState<SortValue>('newest');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [storageModal, setStorageModal] = useState<{ open: boolean; user: AdminUser | null }>({
    open: false,
    user: null,
  });
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [storageForm] = Form.useForm();
  const [createForm] = Form.useForm();

  const { data, isLoading } = useUsers({
    limit: pageSize,
    offset: (page - 1) * pageSize,
    ...(searchQuery.trim() ? { search: searchQuery.trim() } : {}),
    ...(roleFilter ? { role: roleFilter } : {}),
  });

  const createUserMut = useCreateUser();
  const updateRoleMut = useUpdateUserRole();
  const updateStorageMut = useUpdateStorageBase();
  const deleteUserMut = useDeleteUser();

  const users = data?.items ?? [];
  const total = data?.total ?? 0;

  // Client-side sort
  const sortedUsers = [...users].sort((a, b) => {
    switch (sortBy) {
      case 'newest':
        return b.createdAt - a.createdAt;
      case 'oldest':
        return a.createdAt - b.createdAt;
      case 'name-asc':
        return a.username.localeCompare(b.username);
      case 'name-desc':
        return b.username.localeCompare(a.username);
      default:
        return 0;
    }
  });

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await updateRoleMut.mutateAsync({ id: userId, role });
      message.success('Role updated');
    } catch {
      message.error('Failed to update role');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteUserMut.mutateAsync(id);
      message.success('User deleted');
    } catch {
      message.error('Failed to delete user');
    }
  };

  const handleStorageSave = async (values: { baseBytes: number }) => {
    if (!storageModal.user) return;
    try {
      await updateStorageMut.mutateAsync({ id: storageModal.user.id, baseBytes: values.baseBytes });
      message.success('Storage base updated');
      setStorageModal({ open: false, user: null });
    } catch {
      message.error('Failed to update storage');
    }
  };

  const handleCreateUser = async (values: CreateUserInput) => {
    try {
      await createUserMut.mutateAsync(values);
      message.success('User created');
      setCreateModalOpen(false);
      createForm.resetFields();
    } catch {
      message.error('Failed to create user');
    }
  };

  const columns: ColumnsType<AdminUser> = [
    {
      title: 'Username',
      dataIndex: 'username',
      key: 'username',
      render: (text: string, record: AdminUser) => (
        <a onClick={() => navigate(`/admin/users/${record.id}`)} style={{ cursor: 'pointer' }}>
          {text}
        </a>
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
      width: 160,
      render: (_: unknown, record: AdminUser) => (
        <Space>
          <Tooltip title="View">
            <EyeOutlined
              style={{ cursor: 'pointer', color: '#1890ff' }}
              onClick={() => navigate(`/admin/users/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title="Edit Storage Base">
            <EditOutlined
              style={{ cursor: 'pointer', color: '#52c41a' }}
              onClick={() => setStorageModal({ open: true, user: record })}
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
          <Button onClick={() => { setSearchQuery(searchInput); setPage(1); }}>Search</Button>
          <Select
            value={sortBy}
            onChange={(val: SortValue) => setSortBy(val)}
            style={{ width: 150 }}
            options={[
              { label: 'Newest first', value: 'newest' },
              { label: 'Oldest first', value: 'oldest' },
              { label: 'Name A-Z', value: 'name-asc' },
              { label: 'Name Z-A', value: 'name-desc' },
            ]}
          />
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalOpen(true)}>
          Create User
        </Button>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={sortedUsers}
        loading={isLoading}
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

      {/* Create User Modal */}
      <Modal
        title="Create User"
        open={createModalOpen}
        onCancel={() => {
          setCreateModalOpen(false);
          createForm.resetFields();
        }}
        footer={null}
        destroyOnClose
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreateUser}>
          <Form.Item
            name="username"
            label="Username"
            rules={[{ required: true, message: 'Please enter username' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            name="email"
            label="Email"
            rules={[{ required: true, message: 'Please enter email' }]}
          >
            <Input type="email" />
          </Form.Item>
          <Form.Item
            name="password"
            label="Password"
            rules={[{ required: true, message: 'Please enter password' }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="Role" initialValue="user">
            <Select options={ROLES.map((r) => ({ label: r, value: r }))} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={createUserMut.isPending}>
              Create
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit Storage Base Modal */}
      <Modal
        title={`Edit Storage Base - ${storageModal.user?.username}`}
        open={storageModal.open}
        onCancel={() => setStorageModal({ open: false, user: null })}
        footer={null}
        destroyOnClose
      >
        <Form
          form={storageForm}
          layout="vertical"
          onFinish={handleStorageSave}
          initialValues={{ baseBytes: storageModal.user?.baseBytes }}
        >
          <Form.Item
            name="baseBytes"
            label="Base Bytes"
            rules={[{ required: true, message: 'Required' }]}
          >
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={updateStorageMut.isPending}>
              Save
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}