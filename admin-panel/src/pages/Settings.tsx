import { useState } from 'react';
import {
  Table,
  Button,
  Modal,
  Select,
  Input,
  InputNumber,
  Space,
  Spin,
  Result,
  Popconfirm,
  Tag,
  message,
} from 'antd';
import { EditOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  useSystemConfig,
  useUpdateSystemConfig,
  useResetSystemConfig,
} from '../api/admin.hooks';
import type { SystemConfigItem } from '../api/admin';
import type { ColumnsType } from 'antd/es/table';

const UNIT_OPTIONS = [
  { label: 'B', value: 'B' },
  { label: 'KB', value: 'KB' },
  { label: 'MB', value: 'MB' },
  { label: 'GB', value: 'GB' },
  { label: 'TB', value: 'TB' },
];

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function parseBytesInput(val: number, unit: string): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const idx = units.indexOf(unit);
  return String(Math.round(val * Math.pow(1024, idx)));
}

type ConfigType = 'bytes' | 'number' | 'bool' | 'string';

function inferType(item: SystemConfigItem): ConfigType {
  const val = item.value;
  if (val === 'true' || val === 'false') return 'bool';
  if (/^\d+$/.test(val)) {
    const key = item.key.toLowerCase();
    if (
      key.includes('size') ||
      key.includes('quota') ||
      key.includes('bytes') ||
      key.includes('upload') ||
      key.includes('storage')
    ) {
      return 'bytes';
    }
    return 'number';
  }
  return 'string';
}

function getByteUnitFromBytes(b: number): { value: number; unit: string } {
  if (b === 0) return { value: 0, unit: 'MB' };
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return {
    value: parseFloat((b / Math.pow(k, i)).toFixed(2)),
    unit: sizes[i],
  };
}

function formatDisplayValue(item: SystemConfigItem): string {
  const type = inferType(item);
  switch (type) {
    case 'bool':
      return item.value;
    case 'bytes': {
      const b = parseInt(item.value, 10);
      return isNaN(b) ? item.value : formatBytes(b);
    }
    default:
      return item.value;
  }
}

export default function Settings() {
  const { data: configs, isLoading, error } = useSystemConfig();
  const updateConfigMut = useUpdateSystemConfig();
  const resetConfigMut = useResetSystemConfig();

  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editItem, setEditItem] = useState<SystemConfigItem | null>(null);
  const [editType, setEditType] = useState<ConfigType>('string');
  const [editValue, setEditValue] = useState('');
  const [editByteNum, setEditByteNum] = useState(0);
  const [editByteUnit, setEditByteUnit] = useState('MB');

  const openEditModal = (item: SystemConfigItem) => {
    setEditItem(item);
    const type = inferType(item);
    setEditType(type);
    if (type === 'bytes') {
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
      let newValue: string;
      if (editType === 'bytes') {
        newValue = parseBytesInput(editByteNum, editByteUnit);
      } else {
        newValue = editValue;
      }
      await updateConfigMut.mutateAsync({ [editItem.key]: newValue });
      message.success('Configuration updated');
      setEditModalOpen(false);
    } catch {
      message.error('Failed to update configuration');
    }
  };

  const handleReset = async (key?: string) => {
    try {
      await resetConfigMut.mutateAsync(key);
      message.success(key ? `Configuration "${key}" reset` : 'All configurations reset');
    } catch {
      message.error('Failed to reset configuration');
    }
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 60 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: 24 }}>
        <Result
          status="error"
          title="Failed to load system config"
          subTitle={error.message}
        />
      </div>
    );
  }

  if (!configs || configs.length === 0) {
    return (
      <div style={{ padding: 24 }}>
        <Result status="warning" title="No configuration found" />
      </div>
    );
  }

  const columns: ColumnsType<SystemConfigItem> = [
    {
      title: 'Setting',
      key: 'setting',
      render: (_: unknown, record: SystemConfigItem) => (
        <div>
          {record.description && (
            <div style={{ fontWeight: 500 }}>{record.description}</div>
          )}
          <code style={{ fontSize: 12, color: '#999' }}>{record.key}</code>
        </div>
      ),
    },
    {
      title: 'Current Value',
      dataIndex: 'value',
      key: 'value',
      width: 200,
      render: (_: unknown, record: SystemConfigItem) => (
        <code>{formatDisplayValue(record)}</code>
      ),
    },
    {
      title: 'Default Value',
      key: 'defaultValue',
      width: 150,
      render: () => <span style={{ color: '#999' }}>-</span>,
    },
    {
      title: 'Type',
      key: 'type',
      width: 100,
      render: (_: unknown, record: SystemConfigItem) => {
        const type = inferType(record);
        const colorMap: Record<ConfigType, string> = {
          bytes: 'purple',
          number: 'blue',
          bool: 'cyan',
          string: 'green',
        };
        return <Tag color={colorMap[type]}>{type}</Tag>;
      },
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 160,
      render: (_: unknown, record: SystemConfigItem) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            Edit
          </Button>
          <Popconfirm
            title={`Reset "${record.key}"?`}
            description="This will reset to the default value."
            onConfirm={() => handleReset(record.key)}
            okText="Yes"
            cancelText="No"
          >
            <Button type="link" size="small" icon={<ReloadOutlined />}>
              Reset
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          marginBottom: 16,
          alignItems: 'center',
        }}
      >
        <h2 style={{ margin: 0 }}>System Configuration</h2>
        <Popconfirm
          title="Reset all configurations?"
          description="This will reset all settings to their default values."
          onConfirm={() => handleReset()}
          okText="Yes"
          cancelText="No"
        >
          <Button icon={<ReloadOutlined />} loading={resetConfigMut.isPending}>
            Reset All
          </Button>
        </Popconfirm>
      </div>

      <Table
        rowKey="key"
        columns={columns}
        dataSource={configs}
        loading={isLoading}
        pagination={false}
        size="small"
      />

      {/* Edit Modal */}
      <Modal
        title={`Edit: ${editItem?.key}`}
        open={editModalOpen}
        onCancel={() => setEditModalOpen(false)}
        onOk={handleEditSave}
        confirmLoading={updateConfigMut.isPending}
      >
        {editType === 'string' && (
          <Input
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
            placeholder="Enter value"
          />
        )}
        {editType === 'number' && (
          <InputNumber
            style={{ width: '100%' }}
            value={parseInt(editValue, 10) || 0}
            onChange={(v) => setEditValue(String(v ?? 0))}
            min={0}
          />
        )}
        {editType === 'bool' && (
          <Select
            style={{ width: '100%' }}
            value={editValue}
            onChange={(v) => setEditValue(v)}
            options={[
              { label: 'true', value: 'true' },
              { label: 'false', value: 'false' },
            ]}
          />
        )}
        {editType === 'bytes' && (
          <Space style={{ width: '100%' }} align="start">
            <InputNumber
              style={{ flex: 1 }}
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
        {editType === 'bytes' && editByteNum > 0 && (
          <div style={{ marginTop: 8, fontSize: 12, color: '#999' }}>
            Raw bytes: {parseBytesInput(editByteNum, editByteUnit)}
          </div>
        )}
      </Modal>
    </div>
  );
}