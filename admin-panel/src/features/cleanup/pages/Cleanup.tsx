import { useState, useCallback } from 'react';
import {
  Card,
  Col,
  Descriptions,
  Input,
  message,
  Popconfirm,
  Result,
  Row,
  Spin,
  Statistic,
  Tabs,
  Tag,
} from 'antd';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  useCleanupQuery,
  useDeleteUserFile,
  useDeletePhysicalFile,
} from '../../../api/admin.hooks';
import type {
  CleanupQueryPhysicalFile,
  CleanupQueryUserFile,
} from '../../../api/admin';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
}

function formatDate(epoch: number): string {
  return new Date(epoch * 1000).toLocaleString();
}

const LS_PREFIX = 'nd.admin.search';

function loadHistory(mode: string): string[] {
  try {
    const raw = localStorage.getItem(`${LS_PREFIX}.${mode}`);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveHistory(mode: string, value: string) {
  const key = `${LS_PREFIX}.${mode}`;
  const history = loadHistory(mode);
  const updated = [value, ...history.filter((h) => h !== value)].slice(0, 10);
  localStorage.setItem(key, JSON.stringify(updated));
}

export default function Cleanup() {
  const { t } = useTranslation();
  const [mode, setMode] = useState<'slug' | 'hash'>('slug');
  const [input, setInput] = useState('');
  const [history, setHistory] = useState<string[]>(() => loadHistory(mode));
  const queryMutation = useCleanupQuery();
  const deleteUserFileMutation = useDeleteUserFile();
  const deletePhysicalFileMutation = useDeletePhysicalFile();

  const handleSearch = useCallback(
    async (value: string) => {
      const trimmed = value.trim();
      if (!trimmed) return;
      saveHistory(mode, trimmed);
      setHistory(loadHistory(mode));
      await queryMutation.mutateAsync(
        mode === 'slug' ? { slug: trimmed } : { hash: trimmed },
      );
    },
    [mode, queryMutation],
  );

  const handleModeChange = (key: string) => {
    setMode(key as 'slug' | 'hash');
    setInput('');
  };

  const result = queryMutation.data;

  const totalUploads = result?.totalUploads ?? 0;
  const uniqueUsers = result?.uniqueUsers ?? 0;
  const totalSize = result
    ? result.userFiles.reduce((a, b) => a + b.fileSize, 0)
    : 0;

  const handleDeleteUserFile = async (userFileId: number) => {
    try {
      await deleteUserFileMutation.mutateAsync(userFileId);
      message.success(t('cleanup.deleteSuccess'));
      queryMutation.mutate({});
    } catch (err: any) {
      message.error(err?.message || t('cleanup.deleteFailed'));
    }
  };

  const handleDeletePhysicalFile = async (physicalFileId: number) => {
    try {
      await deletePhysicalFileMutation.mutateAsync(physicalFileId);
      message.success(t('cleanup.deleteAllSuccess'));
      queryMutation.mutate({});
    } catch (err: any) {
      message.error(err?.message || t('cleanup.deleteFailed'));
    }
  };

  const userFileColumns: ProColumns<CleanupQueryUserFile>[] = [
    { title: t('cleanup.id'), dataIndex: 'id', hideInSearch: true },
    { title: t('cleanup.slug'), dataIndex: 'slug', ellipsis: true, hideInSearch: true },
    { title: t('cleanup.filename'), dataIndex: 'fileName', hideInSearch: true },
    {
      title: t('cleanup.user'),
      key: 'user',
      hideInSearch: true,
      render: (_, record) => (
        <Link to={`/admin/users/${record.userId}`}>{record.username}</Link>
      ),
    },
    {
      title: t('cleanup.size'),
      dataIndex: 'fileSize',
      hideInSearch: true,
      render: (size: number) => formatBytes(size),
    },
    {
      title: t('cleanup.created'),
      dataIndex: 'createdAt',
      hideInSearch: true,
      render: (v: number) => formatDate(v),
    },
    {
      title: t('cleanup.actions'),
      key: 'actions',
      hideInSearch: true,
      render: (_, record) => (
        <Popconfirm
          title={t('cleanup.deleteConfirm')}
          onConfirm={() => handleDeleteUserFile(record.id)}
        >
          <a style={{ color: 'red', cursor: 'pointer' }}>
            <DeleteOutlined /> {t('cleanup.delete')}
          </a>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <h2>{t('cleanup.title')}</h2>

      <Tabs
        activeKey={mode}
        onChange={handleModeChange}
        items={[
          { key: 'slug', label: t('cleanup.bySlug') },
          { key: 'hash', label: t('cleanup.byHash') },
        ]}
      />

      <Input.Search
        placeholder={
          mode === 'slug' ? t('cleanup.slugPlaceholder') : t('cleanup.hashPlaceholder')
        }
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onSearch={handleSearch}
        enterButton={t('cleanup.search')}
        style={{ marginBottom: 16 }}
      />

      {history.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          {history.map((item) => (
            <Tag
              key={item}
              style={{ cursor: 'pointer', marginBottom: 4 }}
              onClick={() => {
                setInput(item);
                handleSearch(item);
              }}
            >
              {item}
            </Tag>
          ))}
        </div>
      )}

      {queryMutation.isPending && (
        <div style={{ textAlign: 'center', padding: 60 }}>
          <Spin size="large" />
        </div>
      )}

      {queryMutation.isError && (
        <Result
          status="error"
          title={t('cleanup.queryFailed')}
          subTitle={queryMutation.error?.message}
        />
      )}

      {result && (
        <>
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={8}>
              <Card>
                <Statistic title={t('cleanup.totalUploads')} value={totalUploads} />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic title={t('cleanup.uniqueUsers')} value={uniqueUsers} />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic title={t('cleanup.totalSize')} value={formatBytes(totalSize)} />
              </Card>
            </Col>
          </Row>

          {result.physicalFile && (
            <Card
              title={t('cleanup.physicalFile')}
              style={{ marginBottom: 24 }}
              extra={
                <Popconfirm
                  title={t('cleanup.deleteAllConfirm')}
                  onConfirm={() => handleDeletePhysicalFile(result.physicalFile!.id)}
                >
                  <a style={{ color: 'red', cursor: 'pointer' }}>
                    <DeleteOutlined /> {t('cleanup.deleteAll')}
                  </a>
                </Popconfirm>
              }
            >
              <Descriptions column={2} bordered size="small">
                <Descriptions.Item label={t('cleanup.id')}>{result.physicalFile.id}</Descriptions.Item>
                <Descriptions.Item label={t('cleanup.hash')}>{result.physicalFile.fileHash}</Descriptions.Item>
                <Descriptions.Item label={t('cleanup.size')}>{formatBytes(result.physicalFile.fileSize)}</Descriptions.Item>
                <Descriptions.Item label={t('cleanup.mimeType')}>{result.physicalFile.mimeType}</Descriptions.Item>
                <Descriptions.Item label={t('cleanup.storagePath')}>{result.physicalFile.storagePath}</Descriptions.Item>
                <Descriptions.Item label={t('cleanup.onDisk')}>
                  <Tag color={result.physicalFile.fileExists ? 'green' : 'red'}>
                    {result.physicalFile.fileExists ? t('cleanup.yes') : t('cleanup.no')}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>
            </Card>
          )}

          {result.userFiles.length > 0 && (
            <ProTable
              rowKey="id"
              columns={userFileColumns}
              dataSource={result.userFiles}
              pagination={false}
              search={false}
              options={false}
              size="small"
            />
          )}
        </>
      )}
    </div>
  );
}