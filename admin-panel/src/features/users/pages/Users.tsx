import { useCallback, useState } from 'react';
import { Space, Select, Button, Popconfirm, Tooltip } from 'antd';
import { ModalForm, ProFormText, ProFormSelect, ProFormDigit } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import {
  DeleteOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { Input } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { PageContainer } from '../../../components/PageContainer';
import { SearchForm } from '../../../components/SearchForm';
import type { SearchField } from '../../../components/SearchForm';
import { ProTable } from '../../../components/ProTable';
import { useTableUrlState } from '../../../hooks/useTableUrlState';
import { useUsers, useCreateUser, useUpdateUserRole, useUpdateStorageBase, useDeleteUser } from '../../../api/admin.hooks';
import type { AdminUser, CreateUserInput } from '../../../api/admin';
import { formatDateShort, formatBytes } from '../../../utils/format';

const ROLES = ['admin', 'user'];

export default function UsersPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [storageModalOpen, setStorageModalOpen] = useState(false);
  const [storageUser, setStorageUser] = useState<AdminUser | null>(null);

  const createUserMut = useCreateUser();
  const updateRoleMut = useUpdateUserRole();
  const updateStorageMut = useUpdateStorageBase();
  const deleteUserMut = useDeleteUser();

  const { params, setParams, resetParams } = useTableUrlState({
    page: 1,
    pageSize: 20,
    search: '',
    role: '',
  });

  const queryResult = useUsers({
    limit: params.pageSize,
    offset: (params.page - 1) * params.pageSize,
    search: params.search || undefined,
    role: params.role || undefined,
  });

  const handleRoleChange = useCallback(
    (userId: string, role: string) => {
      updateRoleMut.mutate({ id: userId, role });
    },
    [updateRoleMut],
  );

  const handleDelete = useCallback(
    (id: string) => {
      deleteUserMut.mutate(id);
    },
    [deleteUserMut],
  );

  const handleSearch = useCallback(
    (values: Record<string, unknown>) => {
      setParams({
        page: 1,
        search: (values.search as string) || '',
        role: (values.role as string) || '',
      });
    },
    [setParams],
  );

  const handlePageChange = useCallback(
    (page: number, pageSize: number) => {
      setParams({ page, pageSize });
    },
    [setParams],
  );

  const columns: ProColumns<AdminUser>[] = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: t('users.registerMethod'), dataIndex: 'registerMethod', width: 110 },
    {
      title: t('users.username'),
      dataIndex: 'username',
      render: (text: string, record: AdminUser) => (
        <a onClick={() => navigate(`/admin/users/${record.id}`)} style={{ cursor: 'pointer' }}>
          {text}
        </a>
      ),
    },
    { title: t('users.email'), dataIndex: 'email' },
    {
      title: t('users.role'),
      dataIndex: 'role',
      width: 130,
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
      render: (v: number) => formatDateShort(v),
    },
    {
      title: t('users.actions'),
      width: 160,
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

  const searchFields: SearchField[] = [
    {
      name: 'search',
      label: t('users.username'),
      children: <Input placeholder={t('users.search')} allowClear />,
    },
    {
      name: 'role',
      label: t('users.role'),
      children: (
        <Select
          allowClear
          placeholder={t('users.allRoles')}
          options={ROLES.map((r) => ({ label: t(`users.${r}`), value: r }))}
          style={{ width: 160 }}
        />
      ),
    },
  ];

  return (
    <PageContainer
      title={t('users.title')}
      extra={[
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
            render: (props, defaultDoms) => [defaultDoms[1], defaultDoms[0]],
          }}
          onFinish={async (values) => {
            try {
              await createUserMut.mutateAsync(values);
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
    >
      <SearchForm
        fields={searchFields}
        onSearch={handleSearch}
        onReset={resetParams}
        initialValues={{ search: params.search || undefined, role: params.role || undefined }}
      />
      <ProTable<AdminUser>
        rowKey="id"
        columns={columns as never}
        queryResult={queryResult}
        pagination={{ current: params.page, pageSize: params.pageSize }}
        onChange={(pag) => {
          if (pag.current && pag.pageSize) {
            handlePageChange(pag.current, pag.pageSize);
          }
        }}
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