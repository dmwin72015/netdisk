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
  Table,
  Tabs,
  Tag,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import {
  useCleanupQuery,
  useDeleteUserFile,
  useDeletePhysicalFile,
} from '../api/admin.hooks';
import type {
  CleanupQueryPhysicalFile,
  CleanupQueryUserFile,
} from '../api/admin';

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
      message.success('User file deleted');
      queryMutation.mutate({});
    } catch (err: any) {
      message.error(err?.message || 'Failed to delete user file');
    }
  };

  const handleDeletePhysicalFile = async (physicalFileId: number) => {
    try {
      await deletePhysicalFileMutation.mutateAsync(physicalFileId);
      message.success('Physical file deleted');
      queryMutation.mutate({});
    } catch (err: any) {
      message.error(err?.message || 'Failed to delete physical file');
    }
  };

  const userFileColumns: ColumnsType<CleanupQueryUserFile> = [
    { title: 'ID', dataIndex: 'id', key: 'id' },
    { title: 'Slug', dataIndex: 'slug', key: 'slug', ellipsis: true },
    { title: 'Filename', dataIndex: 'fileName', key: 'fileName' },
    {
      title: 'User',
      key: 'user',
      render: (_, record) => (
        <Link to={`/admin/users/${record.userId}`}>{record.username}</Link>
      ),
    },
    {
      title: 'Size',
      dataIndex: 'fileSize',
      key: 'fileSize',
      render: (size: number) => formatBytes(size),
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (v: number) => formatDate(v),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record) => (
        <Popconfirm
          title="Delete this user file?"
          onConfirm={() => handleDeleteUserFile(record.id)}
        >
          <a style={{ color: 'red', cursor: 'pointer' }}>
            <DeleteOutlined /> Delete
          </a>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <h2>File Cleanup</h2>

      <Tabs
        activeKey={mode}
        onChange={handleModeChange}
        items={[
          { key: 'slug', label: 'By Slug' },
          { key: 'hash', label: 'By Hash' },
        ]}
      />

      <Input.Search
        placeholder={
          mode === 'slug' ? 'Enter file slug...' : 'Enter file hash...'
        }
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onSearch={handleSearch}
        enterButton="Search"
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
          title="Query Failed"
          subTitle={queryMutation.error?.message}
        />
      )}

      {result && (
        <>
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={8}>
              <Card>
                <Statistic title="Total Uploads" value={totalUploads} />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic title="Unique Users" value={uniqueUsers} />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic title="Total Size" value={formatBytes(totalSize)} />
              </Card>
            </Col>
          </Row>

          {result.physicalFile && (
            <Card
              title="Physical File"
              style={{ marginBottom: 24 }}
              extra={
                <Popconfirm
                  title="Delete this physical file and ALL associated records?"
                  onConfirm={() => handleDeletePhysicalFile(result.physicalFile!.id)}
                >
                  <a style={{ color: 'red', cursor: 'pointer' }}>
                    <DeleteOutlined /> Delete All
                  </a>
                </Popconfirm>
              }
            >
              <Descriptions column={2} bordered size="small">
                <Descriptions.Item label="ID">{result.physicalFile.id}</Descriptions.Item>
                <Descriptions.Item label="Hash">{result.physicalFile.fileHash}</Descriptions.Item>
                <Descriptions.Item label="Size">{formatBytes(result.physicalFile.fileSize)}</Descriptions.Item>
                <Descriptions.Item label="MIME Type">{result.physicalFile.mimeType}</Descriptions.Item>
                <Descriptions.Item label="Storage Path">{result.physicalFile.storagePath}</Descriptions.Item>
                <Descriptions.Item label="On Disk">
                  <Tag color={result.physicalFile.fileExists ? 'green' : 'red'}>
                    {result.physicalFile.fileExists ? 'Yes' : 'No'}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>
            </Card>
          )}

          {result.userFiles.length > 0 && (
            <Table
              rowKey="id"
              columns={userFileColumns}
              dataSource={result.userFiles}
              pagination={false}
              size="small"
            />
          )}
        </>
      )}
    </div>
  );
}