import { useState, useCallback } from 'react';
import { Button, Card, Col, Descriptions, Input, Popconfirm, Result, Row, Space, Spin, Statistic, Tag, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DeleteOutlined, SearchOutlined, HistoryOutlined, ClearOutlined } from '@ant-design/icons';
import { Link } from 'react-router';
import { PageContainer } from '@/components/PageContainer';
import CopyCell from '@/components/CopyCell';
import { useCleanupQuery, useDeleteUserFile, useDeletePhysicalFile } from '@/api/cleanup';
import type { CleanupQueryUserFile } from '@/api/cleanup';
import { formatBytes, formatDate } from '@/utils/format';

const LS_PREFIX = 'admin.cleanup.search';

function loadHistory(mode: string): string[] {
  try { const raw = localStorage.getItem(`${LS_PREFIX}.${mode}`); return raw ? JSON.parse(raw) : []; } catch { return []; }
}
function saveHistory(mode: string, value: string) {
  const key = `${LS_PREFIX}.${mode}`;
  const history = loadHistory(mode);
  const updated = [value, ...history.filter(h => h !== value)].slice(0, 10);
  localStorage.setItem(key, JSON.stringify(updated));
}
function clearHistory(mode: string) {
  localStorage.removeItem(`${LS_PREFIX}.${mode}`);
}

const MODE_CONFIG = {
  slug: { label: '按 Slug 搜索', placeholder: '输入文件 Slug', key: 'slug' as const },
  hash: { label: '按 Hash 搜索', placeholder: '输入文件 Hash（支持前缀匹配）', key: 'hash' as const },
};

export default function CleanupPage() {
  const [mode, setMode] = useState<'slug' | 'hash'>('slug');
  const [input, setInput] = useState('');
  const [history, setHistory] = useState<string[]>(() => loadHistory(mode));
  const [showHistory, setShowHistory] = useState(false);
  const queryMutation = useCleanupQuery();
  const deleteUserFileMutation = useDeleteUserFile();
  const deletePhysicalFileMutation = useDeletePhysicalFile();

  const handleSearch = useCallback(async (value: string) => {
    const trimmed = value.trim();
    if (!trimmed) return;
    saveHistory(mode, trimmed);
    setHistory(loadHistory(mode));
    setShowHistory(false);
    await queryMutation.mutateAsync(mode === 'slug' ? { slug: trimmed } : { hash: trimmed });
  }, [mode, queryMutation]);

  const switchMode = (key: string) => {
    const newMode = key as 'slug' | 'hash';
    setMode(newMode);
    setInput('');
    setShowHistory(false);
    setHistory(loadHistory(newMode));
  };

  const result = queryMutation.data;
  const totalUploads = result?.totalUploads ?? 0;
  const uniqueUsers = result?.uniqueUsers ?? 0;
  const totalSize = result ? result.userFiles.reduce((a, b) => a + b.fileSize, 0) : 0;

  const handleDeleteUserFile = async (userFileId: number) => {
    try { await deleteUserFileMutation.mutateAsync(userFileId); message.success('已删除'); queryMutation.mutate({} as never); }
    catch (err: unknown) { message.error(err instanceof Error ? err.message : '删除失败'); }
  };

  const handleDeletePhysicalFile = async (physicalFileId: number) => {
    try { await deletePhysicalFileMutation.mutateAsync(physicalFileId); message.success('已删除全部'); queryMutation.mutate({} as never); }
    catch (err: unknown) { message.error(err instanceof Error ? err.message : '删除失败'); }
  };

  const userFileColumns: ColumnsType<CleanupQueryUserFile> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: 'Slug', dataIndex: 'slug', width: 180, ellipsis: true, render: (_, r) => <CopyCell value={r.slug} /> },
    { title: '文件名', dataIndex: 'fileName', width: 200, ellipsis: true },
    { title: '用户', key: 'user', width: 120, render: (_, record) => <Link to={`/users/${record.userId}`}>{record.username}</Link> },
    { title: '大小', dataIndex: 'fileSize', width: 120, render: (_, record) => formatBytes(record.fileSize) },
    { title: '创建时间', dataIndex: 'createdAt', width: 170, render: (_, record) => formatDate(record.createdAt) },
    {
      title: '操作', key: 'actions',
      render: (_, record) => (
        <Space>
          <Popconfirm title="确定删除此用户文件？" onConfirm={() => handleDeleteUserFile(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const currentCfg = MODE_CONFIG[mode];

  return (
    <PageContainer title="清理工具">
      {/* Search Section */}
      <Card hoverable className="mb-6!">
        <div className="flex items-center gap-2 mb-4">
          {(Object.entries(MODE_CONFIG) as [string, typeof MODE_CONFIG.slug][]).map(([key, cfg]) => (
            <Button
              key={key}
              type={mode === key ? 'primary' : 'default'}
              onClick={() => switchMode(key)}
              size="small"
            >
              {cfg.label}
            </Button>
          ))}
        </div>

        <div className="relative">
          <Input.Search
            placeholder={currentCfg.placeholder}
            value={input}
            onChange={e => setInput(e.target.value)}
            onSearch={handleSearch}
            onFocus={() => setShowHistory(true)}
            onBlur={() => setTimeout(() => setShowHistory(false), 200)}
            enterButton="搜索"
            size="large"
          />

          {showHistory && history.length > 0 && (
            <Card
              className="absolute left-0 right-0 top-full z-10 mt-1!"
              styles={{ body: { padding: '8px 12px' } }}
              hoverable
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs text-[#999]"><HistoryOutlined className="mr-1" />搜索历史</span>
                <Button type="link" size="small" icon={<ClearOutlined />} onClick={() => { clearHistory(mode); setHistory([]); }}>
                  清空
                </Button>
              </div>
              <div className="flex flex-wrap gap-1">
                {history.map(item => (
                  <Tag
                    key={item}
                    className="cursor-pointer mb-1!"
                    onClick={() => { setInput(item); handleSearch(item); }}
                    closable
                    onClose={() => {
                      const updated = history.filter(h => h !== item);
                      setHistory(updated);
                      const key = `${LS_PREFIX}.${mode}`;
                      localStorage.setItem(key, JSON.stringify(updated));
                    }}
                  >
                    {item}
                  </Tag>
                ))}
              </div>
            </Card>
          )}
        </div>
      </Card>

      {/* Loading & Error */}
      {queryMutation.isPending && <div className="text-center p-15"><Spin size="large" /></div>}
      {queryMutation.isError && (
        <Card hoverable>
          <Result status="error" title="查询失败" subTitle={queryMutation.error?.message} />
        </Card>
      )}

      {/* Results */}
      {result && (
        <>
          <Row gutter={16} className="mb-6!">
            <Col span={8}><Card hoverable><Statistic title="总上传数" value={totalUploads} /></Card></Col>
            <Col span={8}><Card hoverable><Statistic title="独立用户数" value={uniqueUsers} /></Card></Col>
            <Col span={8}><Card hoverable><Statistic title="总大小" value={formatBytes(totalSize)} /></Card></Col>
          </Row>

          {result.physicalFile && (
            <Card hoverable title="物理文件信息" className="mb-6!"
              extra={
                <Popconfirm title="确定删除整个物理文件及所有关联用户文件？" onConfirm={() => handleDeletePhysicalFile(result.physicalFile!.id)}>
                  <Button type="primary" size="small" danger icon={<DeleteOutlined />}>删除全部</Button>
                </Popconfirm>
              }
            >
              <Descriptions column={2} bordered size="small">
                <Descriptions.Item label="ID">{result.physicalFile.id}</Descriptions.Item>
                <Descriptions.Item label="哈希"><CopyCell value={result.physicalFile.fileHash}><code className="text-xs">{result.physicalFile.fileHash}</code></CopyCell></Descriptions.Item>
                <Descriptions.Item label="大小">{formatBytes(result.physicalFile.fileSize)}</Descriptions.Item>
                <Descriptions.Item label="MIME 类型">{result.physicalFile.mimeType}</Descriptions.Item>
                <Descriptions.Item label="存储路径" span={2}><code className="text-xs">{result.physicalFile.storagePath}</code></Descriptions.Item>
                <Descriptions.Item label="磁盘存在">
                  <Tag color={result.physicalFile.fileExists ? 'green' : 'red'}>{result.physicalFile.fileExists ? '是' : '否'}</Tag>
                </Descriptions.Item>
              </Descriptions>
            </Card>
          )}

          {result.userFiles.length > 0 && (
            <Card hoverable title="用户上传记录">
              <Table<CleanupQueryUserFile>
                rowKey="id"
                columns={userFileColumns}
                dataSource={result.userFiles}
                pagination={false}
                size="small"
                scroll={{ x: 'max-content' }}
              />
            </Card>
          )}
        </>
      )}
    </PageContainer>
  );
}