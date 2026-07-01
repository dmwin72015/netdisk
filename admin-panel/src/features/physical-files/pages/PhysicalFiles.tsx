import { useRef } from "react";
import { Button, Space, Tag } from "antd";
import { ProTable } from "@ant-design/pro-components";
import type { ProColumns, ActionType } from "@ant-design/pro-components";
import { EyeOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router";
import { useTranslation } from "react-i18next";
import dayjs from "dayjs";
import { PageContainer } from "@/components/PageContainer";
import CopyCell from "@/components/CopyCell";
import { fetchPhysicalFiles } from "@/api/admin";
import type { AdminPhysicalFile } from "@/api/admin";
import { formatBytes, formatDate } from "@/utils/format";

const STATUS_OPTIONS = [
  { label: "已完成", value: "completed" },
  { label: "处理中", value: "processing" },
  { label: "失败", value: "failed" },
];

const STATUS_COLORS: Record<string, string> = {
  completed: "green",
  processing: "blue",
  failed: "red",
};

export default function PhysicalFilesPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>(null);

  const columns: ProColumns<AdminPhysicalFile>[] = [
    { title: "ID", dataIndex: "id", width: 60, hideInSearch: true },
    {
      title: t("physicalFiles.slug"),
      dataIndex: "slug",
      width: 140,
      render: (_, r) => <CopyCell value={r.slug} />,
    },
    {
      title: t("physicalFiles.fileHash"),
      dataIndex: "fileHash",
      width: 240,
      render: (_, r) => <CopyCell value={r.fileHash}>{r.fileHash}</CopyCell>,
    },
    {
      title: t("physicalFiles.fileSize"),
      dataIndex: "fileSize",
      width: 100,
      valueType: "digitRange",
      hideInSearch: true,
      render: (_, r) => formatBytes(r.fileSize),
    },
    {
      title: t("physicalFiles.mimeType"),
      dataIndex: "mimeType",
      width: 120,
    },
    {
      title: t("physicalFiles.storagePath"),
      dataIndex: "storagePath",
      width: 200,
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: t("physicalFiles.status"),
      dataIndex: "status",
      width: 100,
      valueType: "select",
      fieldProps: { options: STATUS_OPTIONS },
      render: (_, r) => (
        <Tag color={STATUS_COLORS[r.status] || "default"}>
          {STATUS_OPTIONS.find((o) => o.value === r.status)?.label ?? r.status}
        </Tag>
      ),
    },
    {
      title: t("physicalFiles.referenceCount"),
      key: "referenceCount",
      width: 110,
      hideInSearch: true,
      render: (_, r) => {
        const total = r.userFileCount + r.mediaItemCount;
        return <span>{total}</span>;
      },
    },
    {
      title: t("physicalFiles.createdAt"),
      dataIndex: "createdAt",
      width: 160,
      valueType: "dateTimeRange",
      hideInSearch: true,
      render: (_, r) => formatDate(r.createdAt),
    },
    {
      title: t("physicalFiles.actions"),
      width: 120,
      fixed: "right",
      hideInSearch: true,
      render: (_, r) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/admin/files/physical/${r.id}`)}
          >
            {t("physicalFiles.viewDetail")}
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={t("physicalFiles.title")}>
      <ProTable<AdminPhysicalFile>
        headerTitle={t("physicalFiles.title")}
        rowKey="id"
        actionRef={actionRef}
        columns={columns}
        request={async (params) => {
          const fileSize = params.fileSize as [number, number] | undefined;
          const createdAt = params.createdAt as
            | [dayjs.Dayjs, dayjs.Dayjs]
            | undefined;
          const res = await fetchPhysicalFiles({
            limit: params.pageSize!,
            offset: (params.current! - 1) * params.pageSize!,
            search: (params.slug as string) || undefined,
            status: (params.status as string) || undefined,
            hash_filter: (params.fileHash as string) || undefined,
            mime_filter: (params.mimeType as string) || undefined,
            min_size: fileSize?.[0],
            max_size: fileSize?.[1],
            created_from: createdAt?.[0]?.toISOString(),
            created_to: createdAt?.[1]?.toISOString(),
          });
          return { data: res.items, success: true, total: res.total };
        }}
        search={{ labelWidth: "auto" }}
        scroll={{ x: 1300 }}
        pagination={{
          showSizeChanger: true,
          showTotal: (total) => t("physicalFiles.total_0", { count: total }),
        }}
        size="small"
        options={false}
      />
    </PageContainer>
  );
}
