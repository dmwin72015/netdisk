import { useRef, useState } from "react";
import { Space, Button, Popconfirm, Tag, message } from "antd";
import {
  ProTable,
  ModalForm,
  ProFormText,
  ProFormSelect,
  ProFormDigit,
} from "@ant-design/pro-components";
import type { ProColumns, ActionType } from "@ant-design/pro-components";
import {
  DeleteOutlined,
  EyeOutlined,
  EditOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router";
import { useTranslation } from "react-i18next";
import { PageContainer } from "@/components/PageContainer";
import CopyCell from "@/components/CopyCell";
import {
  useCreateUser,
  useUpdateUserRole,
  useUpdateStorageBase,
  useDeleteUser,
} from "@/api/admin.hooks";
import { fetchUsers } from "@/api/admin";
import type { AdminUser, CreateUserInput } from "@/api/admin";
import { formatDateShort, formatBytes } from "@/utils/format";

const ROLES = ["admin", "user"];

const ROLE_COLORS: Record<string, string> = {
  admin: "red",
  user: "blue",
};

export default function UsersPage() {
  const { t } = useTranslation();
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
      message.error(t("users.deleteFailed"));
    }
  };

  const columns: ProColumns<AdminUser>[] = [
    { title: "ID", dataIndex: "id", width: 80, hideInSearch: true },
    {
      title: t("users.slug"),
      dataIndex: "slug",
      width: 220,
      hideInSearch: true,
      render: (_, r) => <CopyCell value={r.slug} />,
    },
    {
      title: t("users.username"),
      dataIndex: "username",
      ellipsis: true,
      render: (_, record) => (
        <a
          onClick={() => navigate(`/admin/users/${record.id}`)}
          style={{ cursor: "pointer" }}
        >
          {record.username}
        </a>
      ),
    },
    {
      title: t("users.email"),
      dataIndex: "email",
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: t("users.role"),
      dataIndex: "role",
      width: 100,
      valueType: "select",
      fieldProps: {
        options: ROLES.map((r) => ({ label: t(`users.${r}`), value: r })),
      },
      render: (_, record) => {
        const r = String(record.role ?? "").toLowerCase();
        return <Tag color={ROLE_COLORS[r] || "default"}>{t(`users.${r}`)}</Tag>;
      },
    },
    {
      title: t("users.registerMethod"),
      dataIndex: "registerMethod",
      width: 110,
      hideInSearch: true,
    },
    {
      title: t("users.storageLimit"),
      key: "storage",
      width: 180,
      hideInSearch: true,
      render: (_, record) => (
        <span>
          {formatBytes(record.usedBytes)} / {formatBytes(record.totalBytes)}
        </span>
      ),
    },
    {
      title: t("users.registered"),
      dataIndex: "createdAt",
      width: 160,
      hideInSearch: true,
      render: (_, record) => formatDateShort(record.createdAt),
    },
    {
      title: t("users.actions"),
      width: 200,
      fixed: "right",
      hideInSearch: true,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/admin/users/${record.id}`)}
          >
            {t("users.view")}
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setEditUser(record);
              setEditModalOpen(true);
            }}
          >
            {t("common.edit")}
          </Button>
          <Popconfirm
            title={t("users.deleteUser")}
            description={t("users.deleteConfirm")}
            onConfirm={() => handleDelete(record.id)}
            okText={t("common.yes")}
            cancelText={t("common.no")}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {t("common.delete")}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={t("users.title")}>
      <ProTable<AdminUser>
        headerTitle={t("users.title")}
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
        search={{ labelWidth: "auto" }}
        scroll={{ x: "max-content" }}
        pagination={{
          showSizeChanger: true,
          showTotal: (total) => t("users.total_0", { count: total }),
        }}
        size="small"
        options={false}
        toolBarRender={() => [
          <ModalForm<CreateUserInput>
            key="create"
            title={t("users.createUser")}
            trigger={
              <Button type="primary" icon={<PlusOutlined />}>
                {t("users.createButton")}
              </Button>
            }
            submitter={{
              searchConfig: {
                submitText: t("common.save"),
                resetText: t("common.cancel"),
              },
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
              label={t("users.username")}
              rules={[{ required: true, message: t("users.usernameRequired") }]}
            />
            <ProFormText
              name="email"
              label={t("users.email")}
              rules={[{ required: true, message: t("users.emailRequired") }]}
            />
            <ProFormText.Password
              name="password"
              label={t("login.password")}
              rules={[{ required: true, message: t("users.passwordRequired") }]}
            />
            <ProFormSelect
              name="role"
              label={t("users.role")}
              initialValue="user"
              options={ROLES.map((r) => ({ label: t(`users.${r}`), value: r }))}
            />
          </ModalForm>,
        ]}
      />

      <ModalForm
        title={`${t("common.edit")} - ${editUser?.username}`}
        open={editModalOpen}
        onOpenChange={(open) => {
          if (!open) {
            setEditModalOpen(false);
            setEditUser(null);
          }
        }}
        initialValues={{
          role: editUser?.role,
          baseBytes: editUser?.baseBytes,
        }}
        onFinish={async (values) => {
          if (!editUser) return false;
          try {
            const role = values.role as string;
            const baseBytes = values.baseBytes as number;
            if (role !== editUser.role) {
              await updateRoleMut.mutateAsync({ id: editUser.id, role });
            }
            if (baseBytes !== editUser.baseBytes) {
              await updateStorageMut.mutateAsync({
                id: editUser.id,
                baseBytes,
              });
            }
            actionRef.current?.reload();
            return true;
          } catch {
            return false;
          }
        }}
      >
        <ProFormSelect
          name="role"
          label={t("users.role")}
          rules={[{ required: true, message: t("users.role") }]}
          options={ROLES.map((r) => ({ label: t(`users.${r}`), value: r }))}
        />
        <ProFormDigit
          name="baseBytes"
          label={t("users.baseBytes")}
          rules={[{ required: true, message: t("users.baseBytesRequired") }]}
          min={0}
          fieldProps={{ style: { width: "100%" } }}
        />
      </ModalForm>
    </PageContainer>
  );
}
