import { useState } from "react";
import {
  Button,
  Modal,
  Select,
  Input,
  InputNumber,
  Space,
  Popconfirm,
  Tag,
  Card,
} from "antd";
import { ProTable } from "@ant-design/pro-components";
import type { ProColumns } from "@ant-design/pro-components";
import { EditOutlined, ReloadOutlined } from "@ant-design/icons";
import { PageContainer } from "@/components/PageContainer";
import {
  useUpdateSystemConfig,
  useResetSystemConfig,
  fetchSystemConfig,
} from "@/api/system-config";
import type { SystemConfigItem } from "@/api/system-config";
import { formatBytes } from "@/utils/format";

const UNIT_OPTIONS = [
  { label: "B", value: "B" },
  { label: "KB", value: "KB" },
  { label: "MB", value: "MB" },
  { label: "GB", value: "GB" },
  { label: "TB", value: "TB" },
];

function parseBytesInput(val: number, unit: string): string {
  const units = ["B", "KB", "MB", "GB", "TB"];
  const idx = units.indexOf(unit);
  return String(Math.round(val * Math.pow(1024, idx)));
}

type ConfigType = "bytes" | "number" | "bool" | "string";

function inferType(item: SystemConfigItem): ConfigType {
  const val = item.value;
  if (val === "true" || val === "false") return "bool";
  if (/^\d+$/.test(val)) {
    const key = item.key.toLowerCase();
    if (
      key.includes("size") ||
      key.includes("quota") ||
      key.includes("bytes") ||
      key.includes("upload") ||
      key.includes("storage")
    )
      return "bytes";
    return "number";
  }
  return "string";
}

function getByteUnitFromBytes(b: number): { value: number; unit: string } {
  if (b === 0) return { value: 0, unit: "MB" };
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return { value: parseFloat((b / Math.pow(k, i)).toFixed(2)), unit: sizes[i] };
}

function formatDisplayValue(item: { value: unknown }): string {
  const sv = String(item.value ?? "");
  if (item.value === true || item.value === false) return String(item.value);
  if (/^\d+$/.test(sv)) {
    const b = parseInt(sv, 10);
    if (/^\d{5,}$/.test(sv) || isNaN(b)) return isNaN(b) ? sv : formatBytes(b);
    return sv;
  }
  return sv;
}

const TYPE_LABELS: Record<ConfigType, string> = {
  bytes: "字节",
  number: "数字",
  bool: "布尔",
  string: "字符串",
};
const TYPE_COLORS: Record<ConfigType, string> = {
  bytes: "purple",
  number: "blue",
  bool: "cyan",
  string: "green",
};

export default function SettingsPage() {
  const updateConfigMut = useUpdateSystemConfig();
  const resetConfigMut = useResetSystemConfig();
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editItem, setEditItem] = useState<SystemConfigItem | null>(null);
  const [editType, setEditType] = useState<ConfigType>("string");
  const [editValue, setEditValue] = useState("");
  const [editByteNum, setEditByteNum] = useState(0);
  const [editByteUnit, setEditByteUnit] = useState("MB");

  const openEditModal = (item: SystemConfigItem) => {
    setEditItem(item);
    const type = inferType(item);
    setEditType(type);
    if (type === "bytes") {
      const b = parseInt(item.value, 10);
      const parsed = getByteUnitFromBytes(isNaN(b) ? 0 : b);
      setEditByteNum(parsed.value);
      setEditByteUnit(parsed.unit);
    } else {
      setEditValue(item.value);
    }
    setEditModalOpen(true);
  };

  const handleEditSave = async () => {
    if (!editItem) return;
    try {
      const newValue =
        editType === "bytes"
          ? parseBytesInput(editByteNum, editByteUnit)
          : editValue;
      await updateConfigMut.mutateAsync({ [editItem.key]: newValue });
      setEditModalOpen(false);
    } catch {
      /* handled by hook */
    }
  };

  const columns: ProColumns<SystemConfigItem>[] = [
    {
      title: "设置项",
      key: "setting",
      hideInSearch: true,
      render: (_, record) => (
        <div>
          {record.description && (
            <div className="font-medium">{record.description}</div>
          )}
          <code className="text-xs text-[#999]">{record.key}</code>
        </div>
      ),
    },
    {
      title: "当前值",
      dataIndex: "value",
      width: 200,
      hideInSearch: true,
      render: (_, record) => <code>{formatDisplayValue(record)}</code>,
    },
    {
      title: "默认值",
      width: 150,
      hideInSearch: true,
      render: (_, record) => {
        const dv = record.defaultValue;
        const hasDefault = dv !== undefined && dv !== null;
        const formatted = hasDefault ? formatDisplayValue({ value: dv }) : "-";
        return (
          <span className={hasDefault ? "text-[#999]" : "text-[#ccc]"}>
            {formatted}
          </span>
        );
      },
    },
    {
      title: "类型",
      width: 100,
      hideInSearch: true,
      render: (_, record) => {
        const type = inferType(record);
        return <Tag color={TYPE_COLORS[type]}>{TYPE_LABELS[type]}</Tag>;
      },
    },
    {
      title: "操作",
      width: 160,
      hideInSearch: true,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要重置此配置吗？"
            onConfirm={() => resetConfigMut.mutate(record.key)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" icon={<ReloadOutlined />}>
              重置
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title="系统设置" extra={[]}>
      <Card
        size="small"
        variant="borderless"
        extra={
          <Popconfirm
            key="resetAll"
            title="确定要重置所有配置吗？"
            onConfirm={() => resetConfigMut.mutate(undefined)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              icon={<ReloadOutlined />}
              loading={resetConfigMut.isPending}
            >
              全部重置
            </Button>
          </Popconfirm>
        }
      >
        <ProTable<SystemConfigItem>
          rowKey="key"
          columns={columns}
          request={async () => {
            const data = await fetchSystemConfig();
            return { data, success: true, total: data.length };
          }}
          pagination={false}
          size="small"
          options={false}
          search={false}
          scroll={{ x: "max-content" }}
        />
      </Card>

      <Modal
        title={`编辑: ${editItem?.key}`}
        open={editModalOpen}
        onCancel={() => setEditModalOpen(false)}
        onOk={handleEditSave}
        confirmLoading={updateConfigMut.isPending}
      >
        {editType === "string" && (
          <Input
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
            placeholder="输入值"
          />
        )}
        {editType === "number" && (
          <InputNumber
            className="w-full"
            value={parseInt(editValue, 10) || 0}
            onChange={(v) => setEditValue(String(v ?? 0))}
            min={0}
          />
        )}
        {editType === "bool" && (
          <Select
            className="w-full"
            value={editValue}
            onChange={(v) => setEditValue(v)}
            options={[
              { label: "true", value: "true" },
              { label: "false", value: "false" },
            ]}
          />
        )}
        {editType === "bytes" && (
          <Space className="w-full" align="start">
            <InputNumber
              className="flex-1"
              value={editByteNum}
              onChange={(v) => setEditByteNum(v ?? 0)}
              min={0}
              precision={2}
            />
            <Select
              value={editByteUnit}
              onChange={(v) => setEditByteUnit(v)}
              style={{ width: 80 }}
              options={UNIT_OPTIONS}
            />
          </Space>
        )}
        {editType === "bytes" && editByteNum > 0 && (
          <div className="mt-2 text-xs text-[#999]">
            原始值: {parseBytesInput(editByteNum, editByteUnit)} 字节
          </div>
        )}
      </Modal>
    </PageContainer>
  );
}
