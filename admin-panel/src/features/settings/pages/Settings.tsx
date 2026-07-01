import { useState } from 'react';
import {
  Button,
  Modal,
  Select,
  Input,
  InputNumber,
  Space,
  Popconfirm,
  Tag,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { EditOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { PageContainer } from '../../../components/PageContainer';
import { ProTable } from '../../../components/ProTable';
import { useTableUrlState } from '../../../hooks/useTableUrlState';
import { useSystemConfig, useUpdateSystemConfig, useResetSystemConfig } from '../../../api/admin.hooks';
import type { SystemConfigItem } from '../../../api/admin';
import { formatBytes, formatDate } from '../../../utils/format';

const UNIT_OPTIONS = [
  { label: 'B', value: 'B' },
  { label: 'KB', value: 'KB' },
  { label: 'MB', value: 'MB' },
  { label: 'GB', value: 'GB' },
  { label: 'TB', value: 'TB' },
];

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

export default function SettingsPage() {
  const { t } = useTranslation();
  const updateConfigMut = useUpdateSystemConfig();
  const resetConfigMut = useResetSystemConfig();

  const [editModalOpen, setEditModalOpen] = useState(false);
  const [editItem, setEditItem] = useState<SystemConfigItem | null>(null);
  const [editType, setEditType] = useState<ConfigType>('string');
  const [editValue, setEditValue] = useState('');
  const [editByteNum, setEditByteNum] = useState(0);
  const [editByteUnit, setEditByteUnit] = useState('MB');

  const queryResult = useSystemConfig();
  const { refetch } = queryResult;

  const { params, setParams } = useTableUrlState({
    page: 1,
    pageSize: 100,
  });

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

  // Map useQuery result to ListQueryResult-compatible shape
  const listQueryResult = {
    data: (queryResult.data as SystemConfigItem[]) ?? [],
    total: (queryResult.data as SystemConfigItem[])?.length ?? 0,
    isLoading: queryResult.isLoading,
    isFetching: queryResult.isFetching,
    isError: queryResult.isError,
    error: queryResult.error as Error | null,
    refetch: () => queryResult.refetch(),
  };

  const columns: ColumnsType<SystemConfigItem> = [
    {
      title: t('settings.setting'),
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
      title: t('settings.currentValue'),
      dataIndex: 'value',
      width: 200,
      render: (_: unknown, record: SystemConfigItem) => (
        <code>{formatDisplayValue(record)}</code>
      ),
    },
    {
      title: t('settings.defaultValue'),
      width: 150,
      render: () => <span style={{ color: '#999' }}>-</span>,
    },
    {
      title: t('settings.type'),
      width: 100,
      render: (_: unknown, record: SystemConfigItem) => {
        const type = inferType(record);
        const colorMap: Record<ConfigType, string> = {
          bytes: 'purple',
          number: 'blue',
          bool: 'cyan',
          string: 'green',
        };
        return <Tag color={colorMap[type]}>{t(`settings.${type}`)}</Tag>;
      },
    },
    {
      title: t('settings.actions'),
      width: 160,
      render: (_: unknown, record: SystemConfigItem) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            {t('settings.edit')}
          </Button>
          <Popconfirm
            title={t('settings.resetConfirm')}
            description={t('settings.resetDescription')}
            onConfirm={() => resetConfigMut.mutate(record.key)}
            okText={t('common.yes')}
            cancelText={t('common.no')}
          >
            <Button type="link" size="small" icon={<ReloadOutlined />}>
              {t('settings.reset')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer
      title={t('settings.title')}
      extra={[
        <Popconfirm
          key="resetAll"
          title={t('settings.resetAllConfirm')}
          description={t('settings.resetAllDescription')}
          onConfirm={() => resetConfigMut.mutate()}
          okText={t('common.yes')}
          cancelText={t('common.no')}
        >
          <Button icon={<ReloadOutlined />} loading={resetConfigMut.isPending}>
            {t('settings.resetAll')}
          </Button>
        </Popconfirm>,
      ]}
    >
      <ProTable<SystemConfigItem>
        rowKey="key"
        columns={columns}
        queryResult={listQueryResult}
        pagination={false as never}
        size="small"
      />

      {/* Edit Modal */}
      <Modal
        title={`${t('settings.edit')}: ${editItem?.key}`}
        open={editModalOpen}
        onCancel={() => setEditModalOpen(false)}
        onOk={() => {
          if (!editItem) return;
          let newValue: string;
          if (editType === 'bytes') {
            newValue = parseBytesInput(editByteNum, editByteUnit);
          } else {
            newValue = editValue;
          }
          updateConfigMut.mutate(
            { [editItem.key]: newValue },
            { onSettled: () => setEditModalOpen(false) },
          );
        }}
        confirmLoading={updateConfigMut.isPending}
      >
        {editType === 'string' && (
          <Input
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
            placeholder={t('settings.enterValue')}
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
              { label: t('settings.true'), value: 'true' },
              { label: t('settings.false'), value: 'false' },
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
            {t('settings.rawBytes', { bytes: parseBytesInput(editByteNum, editByteUnit) })}
          </div>
        )}
      </Modal>
    </PageContainer>
  );
}