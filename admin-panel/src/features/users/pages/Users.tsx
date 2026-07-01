import { useRef, useState } from 'react';
import { Space, Select, Button, Popconfirm, message } from 'antd';
import { ProTable, ModalForm, ProFormText, ProFormSelect, ProFormDigit } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import {
  DeleteOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { PageContainer } from '@/components/PageContainer';
import {
  useCreateUser,
  useUpdateUserRole,
  useUpdateStorageBase,
  useDeleteUser,
} from '@/api/admin.hooks';
import { fetchUsers } from '@/api/admin';
import type { AdminUser, CreateUserInput } from '@/api/admin';
import { formatDateShort, formatBytes } from '@/utils/format';

const ROLES = ['admin', 'user'];

export default function UsersPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);
  const [storageModalOpen, setStorageModalOpen] = useState(false);
  const [storageUser, setStorageUser] = useState<AdminUser | null>(null);

  const createUserMut = useCreateUser();
  const updateRoleMut = useUpdateUserRole();
  const updateStorageMut = useUpdateStorageBase();
  const deleteUserMut = useDeleteUser();

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await updateRoleMut.mutateAsync({ id: userId, role });
      actionRef.current?.reload();
    } catch {
      message.error(t('users.roleFailed'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteUserMut.mutateAsync(id);
      actionRef.current?.reload();
    } catch {
      message.error(t('users.deleteFailed'));
    }
  };

  const columns: ProColumns<AdminUser>[] = [
    { title: 'ID', dataIndex: 'id', width: 80, hideInSearch: true },
    {
      title: t('users.username'),
      dataIndex: 'username',
      render: (_, record) => (
        <a onClick={() => navigate(`/admin/users/${record.id}`)} style={{ cursor: 'pointer' }}>
          {record.username}
        </a>
      ),
    },
    { title: t('users.email'), dataIndex: 'email', hideInSearch: true },
    {
      title: t('users.role'),
      dataIndex: 'role',
      width: 130,
      valueType: 'select',
      fieldProps: { options: ROLES.map((r) => ({ label: t(`users.${r}`), value: r })) },
      render: (role, record) => (
        <Select
          value={role as string}
          size="small"
          style={{ width: 120 }}
          onChange={(val) => handleRoleChange(record.id, val as string)}
          options={ROLES.map((r) => ({ label: t(`users.${r}`), value: r }))}
        />
      ),
    },
    {
      title: t('users.registerMethod'),
      dataIndex: 'registerMethod',
      width: 110,
      hideInSearch: true,
    },
    {
      title: t('users.storageLimit'),
      key: 'storage',
      width: 130,
      hideInSearch: true,
      render: (_, record) => (
        <Tooltip title={`${t('users.total')}: ${formatBytes(record.totalBytes)}`}>
          {formatBytes(record.usedBytes)}
        </Tooltip>
      ),
    },
    {
      title: t('users.registered'),
      dataIndex: 'createdAt',
      width: 120,
      hideInSearch: true,
      render: (_, record) => formatDateShort(record.createdAt),
    },
    {
      title: t('users.actions'),
      width: 160,
      hideInSearch: true,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => navigate(`/admin/users/${record.id}`)}>
            {t('users.view')}
          </Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => { setStorageUser(record); setStorageModalOpen(true); }}>
            {t('users.editStorageBase')}
          </Button>
          <Popconfirm
            title={t('users.deleteUser')}
            description={t('users.deleteConfirm')}
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
    <PageContainer title={t('users.title')}>
      <ProTable<AdminUser>
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
        pagination={{
          showSizeChanger: true,
          showTotal: (total) => t('users.total_0', { count: total }),
        }}
        size="small"
        options={false}
        toolBarRender={() => [
          <ModalForm<CreateUserInput>
            key="create"
            title={t('users.createUser')}
            trigger={
              <Button type="primary" icon={<PlusOutlined />}>
                {t('users.createButton')}
              </Button>
            }
            submitter={{
              searchConfig: { submitText: t('common.save'), resetText: t('common.cancel') },
              render: (_p, defaultDoms) => [defaultDoms[1], defaultDoms[0]],
            }}
            onFinish={async (values) => {
              try {
                await createUserMut.mutateAsync(values);
                actionRef.current?.reload();
                return true;
              } catch {
                return false;
              }
            }}
          >
            <ProFormText
              name="username"
              label={t('users.username')}
              rules={[{ required: true, message: t('users.usernameRequired') }]}
            />
            <ProFormText
              name="email"
              label={t('users.email')}
              rules={[{ required: true, message: t('users.emailRequired') }]}
            />
            <ProFormText.Password
              name="password"
              label={t('login.password')}
              rules={[{ required: true, message: t('users.passwordRequired') }]}
            />
            <ProFormSelect
              name="role"
              label={t('users.role')}
              initialValue="user"
              options={ROLES.map((r) => ({ label: t(`users.${r}`), value: r }))}
            />
          </ModalForm>,
        ]}
      />

      <ModalForm
        title={`${t('users.editStorageBase')} - ${storageUser?.username}`}
        open={storageModalOpen}
        onOpenChange={(open) => {
          if (!open) {
            setStorageModalOpen(false);
            setStorageUser(null);
          }
        }}
        initialValues={{ baseBytes: storageUser?.baseBytes }}
        onFinish={async (values) => {
          if (!storageUser) return false;
          try {
            await updateStorageMut.mutateAsync({
              id: storageUser.id,
              baseBytes: values.baseBytes as number,
            });
            actionRef.current?.reload();
            return true;
          } catch {
            return false;
          }
        }}
      >
        <ProFormDigit
          name="baseBytes"
          label={t('users.baseBytes')}
          rules={[{ required: true, message: t('users.baseBytesRequired') }]}
          min={0}
          fieldProps={{ style: { width: '100%' } }}
        />
      </ModalForm>
    </PageContainer>
  );
}