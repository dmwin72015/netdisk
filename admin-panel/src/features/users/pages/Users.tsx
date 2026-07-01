import { useRef, useState } from 'react';
import { Space, Select, Button, message, Popconfirm, Tooltip } from 'antd';
import { ProTable, ModalForm, ProFormText, ProFormSelect, ProFormDigit } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import {
  DeleteOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  useCreateUser,
  useUpdateUserRole,
  useUpdateStorageBase,
  useDeleteUser,
} from '../../../api/admin.hooks';
import { fetchUsers } from '../../../api/admin';
import type { AdminUser, CreateUserInput } from '../../../api/admin';

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

export default function Users() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>();
  const [storageModalOpen, setStorageModalOpen] = useState(false);
  const [storageUser, setStorageUser] = useState<AdminUser | null>(null);

  const createUserMut = useCreateUser();
  const updateRoleMut = useUpdateUserRole();
  const updateStorageMut = useUpdateStorageBase();
  const deleteUserMut = useDeleteUser();

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      await updateRoleMut.mutateAsync({ id: userId, role });
      message.success(t('users.roleUpdated'));
      actionRef.current?.reload();
    } catch {
      message.error(t('users.roleFailed'));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteUserMut.mutateAsync(id);
      message.success(t('users.userDeleted'));
      actionRef.current?.reload();
    } catch {
      message.error(t('users.deleteFailed'));
    }
  };

  const columns: ProColumns<AdminUser>[] = [
    { title: 'ID', dataIndex: 'id', width: 80, hideInSearch: true },
    { title: t('users.registerMethod'), dataIndex: 'registerMethod', width: 110, hideInSearch: true },
    {
      title: t('users.username'),
      dataIndex: 'username',
      render: (text: string, record: AdminUser) => (
        <a onClick={() => navigate(`/admin/users/${record.id}`)} style={{ cursor: 'pointer' }}>
          {text}
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
      render: (role: string, record: AdminUser) => (
        <Select
          value={role}
          size="small"
          style={{ width: 120 }}
          onChange={(val) => handleRoleChange(record.id, val)}
          options={ROLES.map((r) => ({ label: t(`users.${r}`), value: r }))}
        />
      ),
    },
    {
      title: t('users.storageLimit'),
      key: 'storage',
      width: 130,
      hideInSearch: true,
      render: (_: unknown, record: AdminUser) => (
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
      render: (v: number) => formatDate(v),
    },
    {
      title: t('users.actions'),
      width: 160,
      hideInSearch: true,
      render: (_: unknown, record: AdminUser) => (
        <Space>
          <Tooltip title={t('users.view')}>
            <EyeOutlined
              style={{ cursor: 'pointer', color: '#1890ff' }}
              onClick={() => navigate(`/admin/users/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title={t('users.editStorageBase')}>
            <EditOutlined
              style={{ cursor: 'pointer', color: '#52c41a' }}
              onClick={() => {
                setStorageUser(record);
                setStorageModalOpen(true);
              }}
            />
          </Tooltip>
          <Popconfirm
            title={t('users.deleteUser')}
            description={t('users.deleteConfirm')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('common.yes')}
            cancelText={t('common.no')}
          >
            <DeleteOutlined style={{ cursor: 'pointer', color: '#ff4d4f' }} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
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
              render: (props, defaultDoms) => [
                defaultDoms[1],
                defaultDoms[0],
              ],
            }}
            onFinish={async (values) => {
              try {
                await createUserMut.mutateAsync(values);
                message.success(t('users.createSuccess') || 'User created');
                actionRef.current?.reload();
                return true;
              } catch {
                message.error(t('users.createFailed'));
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
              baseBytes: values.baseBytes,
            });
            message.success(t('users.storageUpdated'));
            actionRef.current?.reload();
            return true;
          } catch {
            message.error(t('users.storageUpdateFailed'));
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
    </>
  );
}