import { useRef, useState } from 'react';
import { Space, Button, Popconfirm, Tag, message } from 'antd';
import { ProTable, ModalForm, ProFormText, ProFormSelect, ProFormDigit } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { DeleteOutlined, EyeOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router';
import { PageContainer } from '@/components/PageContainer';
import CopyCell from '@/components/CopyCell';
import { useCreateUser, useUpdateUserRole, useUpdateStorageBase, useDeleteUser } from '@/api/admin.hooks';
import { fetchUsers } from '@/api/admin';
import type { AdminUser, CreateUserInput } from '@/api/admin';
import { formatDateShort, formatBytes } from '@/utils/format';

const ROLES = ['admin', 'user'];
const ROLE_COLORS: Record<string, string> = { admin: 'red', user: 'blue' };
const ROLE_LABELS: Record<string, string> = { admin: '管理员', user: '普通用户' };

export default function UsersPage() {
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editUser, setEditUser] = useState<AdminUser | null>(null);

  const createUserMut = useCreateUser();
  const updateRoleMut = useUpdateUserRole();
  const updateStorageMut = useUpdateStorageBase();
  const deleteUserMut = useDeleteUser();

  const handleDelete = async (id: string) => {
    try {
      await deleteUserMut.mutateAsync(id);
      actionRef.current?.reload();
    } catch {
      message.error('删除失败');
    }
  };

  const columns: ProColumns<AdminUser>[] = [
    { title: 'ID', dataIndex: 'id', width: 80, hideInSearch: true },
    { title: 'Slug', dataIndex: 'slug', width: 220, hideInSearch: true, render: (_, r) => <CopyCell value={r.slug} /> },
    {
      title: '用户名', dataIndex: 'username', ellipsis: true,
      render: (_, record) => (
        <a onClick={() => navigate(`/users/${record.id}`)} className="cursor-pointer">{record.username}</a>
      ),
    },
    { title: '邮箱', dataIndex: 'email', ellipsis: true, hideInSearch: true },
    {
      title: '角色', dataIndex: 'role', width: 100, valueType: 'select',
      fieldProps: { options: ROLES.map(r => ({ label: ROLE_LABELS[r], value: r })) },
      render: (_, record) => <Tag color={ROLE_COLORS[record.role] || 'default'}>{ROLE_LABELS[record.role] || record.role}</Tag>,
    },
    { title: '注册方式', dataIndex: 'registerMethod', width: 110, hideInSearch: true },
    {
      title: '存储配额', key: 'storage', width: 180, hideInSearch: true,
      render: (_, record) => <span>{formatBytes(record.usedBytes)} / {formatBytes(record.totalBytes)}</span>,
    },
    {
      title: '注册时间', dataIndex: 'createdAt', width: 160, hideInSearch: true,
      render: (_, record) => formatDateShort(record.createdAt),
    },
    {
      title: '操作', width: 200, fixed: 'right', hideInSearch: true,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/users/${record.id}`)}>查看</Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => { setEditUser(record); setEditModalOpen(true); }}>编辑</Button>
          <Popconfirm title="删除用户" description="确定要删除此用户吗？" onConfirm={() => handleDelete(record.id)} okText="确定" cancelText="取消">
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title="用户管理">
      <ProTable<AdminUser>
        headerTitle="用户列表"
        rowKey="id"
        actionRef={actionRef}
        columns={columns}
        request={async (params) => {
          const res = await fetchUsers({
            limit: params.pageSize!,
            offset: (params.current! - 1) * params.pageSize!,
            search: (params.username as string) || undefined,
            role: (params.role as string) || undefined,
          });
          return { data: res.items, success: true, total: res.total };
        }}
        search={{ labelWidth: 'auto' }}
        scroll={{ x: 'max-content' }}
        pagination={{ showSizeChanger: true, showTotal: (total) => `共 ${total} 条` }}
        size="small"
        options={false}
        toolBarRender={() => [
          <ModalForm<CreateUserInput>
            key="create"
            title="创建用户"
            trigger={<Button type="primary" icon={<PlusOutlined />}>新建用户</Button>}
            submitter={{ searchConfig: { submitText: '保存', resetText: '取消' }, render: (_p, defaultDoms) => [defaultDoms[1], defaultDoms[0]] }}
            onFinish={async (values) => {
              try { await createUserMut.mutateAsync(values); actionRef.current?.reload(); return true; } catch { return false; }
            }}
          >
            <ProFormText name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]} />
            <ProFormText name="email" label="邮箱" rules={[{ required: true, message: '请输入邮箱' }]} />
            <ProFormText.Password name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]} />
            <ProFormSelect name="role" label="角色" initialValue="user" options={ROLES.map(r => ({ label: ROLE_LABELS[r], value: r }))} />
          </ModalForm>,
        ]}
      />

      <ModalForm
        title={`编辑 - ${editUser?.username}`}
        open={editModalOpen}
        onOpenChange={(open) => { if (!open) { setEditModalOpen(false); setEditUser(null); } }}
        initialValues={{ role: editUser?.role, baseBytes: editUser?.baseBytes }}
        onFinish={async (values) => {
          if (!editUser) return false;
          try {
            const role = values.role as string;
            const baseBytes = values.baseBytes as number;
            if (role !== editUser.role) await updateRoleMut.mutateAsync({ id: editUser.id, role });
            if (baseBytes !== editUser.baseBytes) await updateStorageMut.mutateAsync({ id: editUser.id, baseBytes });
            actionRef.current?.reload();
            return true;
          } catch { return false; }
        }}
      >
        <ProFormSelect name="role" label="角色" rules={[{ required: true }]} options={ROLES.map(r => ({ label: ROLE_LABELS[r], value: r }))} />
        <ProFormDigit name="baseBytes" label="基础存储配额(字节)" rules={[{ required: true }]} min={0} fieldProps={{ style: { width: '100%' } }} />
      </ModalForm>
    </PageContainer>
  );
}
