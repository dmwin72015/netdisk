import { useState, useCallback } from 'react';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Input,
  Popconfirm,
  Result,
  Row,
  Space,
  Spin,
  Statistic,
  Tabs,
  Tag,
  Table,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router';
import { useTranslation } from 'react-i18next';
import { PageContainer } from '@/components/PageContainer';
import CopyCell from '@/components/CopyCell';
import {
  useCleanupQuery,
  useDeleteUserFile,
  useDeletePhysicalFile,
} from '@/api/admin.hooks';
import type {
  CleanupQueryUserFile,
} from '@/api/admin';
import { formatBytes, formatDate } from '@/utils/format';

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

export default function CleanupPage() {
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
      queryMutation.mutate({} as never);
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : '';
      message.error(errMsg || t('cleanup.deleteFailed'));
    }
  };

  const handleDeletePhysicalFile = async (physicalFileId: number) => {
    try {
      await deletePhysicalFileMutation.mutateAsync(physicalFileId);
      message.success(t('cleanup.deleteAllSuccess'));
      queryMutation.mutate({} as never);
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : '';
      message.error(errMsg || t('cleanup.deleteFailed'));
    }
  };

  const userFileColumns: ColumnsType<CleanupQueryUserFile> = [
    { title: t('cleanup.id'), dataIndex: 'id', width: 80 },
    { title: t('cleanup.slug'), dataIndex: 'slug', ellipsis: true, render: (_, r) => <CopyCell value={r.slug} /> },
    { title: t('cleanup.filename'), dataIndex: 'fileName', ellipsis: true },
    {
      title: t('cleanup.user'),
      key: 'user',
      render: (_, record) => (
        <Link to={`/admin/users/${record.userId}`}>{record.username}</Link>
      ),
    },
    {
      title: t('cleanup.size'),
      dataIndex: 'fileSize',
      render: (_, record) => formatBytes(record.fileSize),
    },
    {
      title: t('cleanup.created'),
      dataIndex: 'createdAt',
      render: (_, record) => formatDate(record.createdAt),
    },
    {
      title: t('cleanup.actions'),
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Popconfirm
            title={t('cleanup.deleteConfirm')}
            onConfirm={() => handleDeleteUserFile(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {t('cleanup.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer title={t('cleanup.title')}>
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
                  <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                    {t('cleanup.deleteAll')}
                  </Button>
                </Popconfirm>
              }
            >
              <Descriptions column={2} bordered size="small">
                <Descriptions.Item label={t('cleanup.id')}>{result.physicalFile.id}</Descriptions.Item>
                <Descriptions.Item label={t('cleanup.hash')}><CopyCell value={result.physicalFile.fileHash}><code style={{fontSize:12}}>{result.physicalFile.fileHash}</code></CopyCell></Descriptions.Item>
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
            <Table<CleanupQueryUserFile>
              rowKey="id"
              columns={userFileColumns}
              dataSource={result.userFiles}
              pagination={false}
              size="small"
            />
          )}
        </>
      )}
    </PageContainer>
  );
}