import { useParams, Link } from 'react-router';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  Card,
  Col,
  Result,
  Row,
  Space,
  Spin,
  Statistic,
  Table,
  Tag,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import { ProDescriptions } from '@ant-design/pro-components';
import { PageContainer } from '@/components/PageContainer';
import CopyCell from '@/components/CopyCell';
import { fetchPhysicalFileDetail } from '@/api/admin';
import type { CleanupQueryUserFile } from '@/api/admin';
import { formatBytes, formatDate } from '@/utils/format';

const STATUS_COLORS: Record<string, string> = {
  completed: 'green',
  processing: 'blue',
  failed: 'red',
};

export default function PhysicalFileDetailPage() {
  const { t } = useTranslation();
  const { slug } = useParams<{ slug: string }>();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['admin', 'physical-file-detail', slug],
    queryFn: () => fetchPhysicalFileDetail(Number(slug)),
    enabled: !!slug,
  });

  const userFileColumns: ColumnsType<CleanupQueryUserFile> = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: t('cleanup.slug'), dataIndex: 'slug', ellipsis: true },
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
  ];

  if (isLoading) {
    return (
      <PageContainer title={t('physicalFiles.detailTitle')}>
        <div style={{ textAlign: 'center', padding: 60 }}>
          <Spin size="large" />
        </div>
      </PageContainer>
    );
  }

  if (isError || !data) {
    return (
      <PageContainer title={t('physicalFiles.detailTitle')}>
        <Result
          status="error"
          title={t('physicalFiles.loadFailed')}
          subTitle={error instanceof Error ? error.message : ''}
          extra={
            <Link to="/admin/files/physical">
              <Button icon={<ArrowLeftOutlined />}>{t('common.back')}</Button>
            </Link>
          }
        />
      </PageContainer>
    );
  }

  const { physicalFile, userFiles, totalUploads, uniqueUsers } = data;
  const totalSize = userFiles.reduce((a, b) => a + b.fileSize, 0);

  return (
    <PageContainer
      title={t('physicalFiles.detailTitle')}
      extra={
        <Link to="/admin/files/physical">
          <Button icon={<ArrowLeftOutlined />}>{t('physicalFiles.backToList')}</Button>
        </Link>
      }
    >
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={6}>
          <Card>
            <Statistic title={t('physicalFiles.totalUploads')} value={totalUploads} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={t('physicalFiles.uniqueUsers')} value={uniqueUsers} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={t('physicalFiles.totalSize')} value={formatBytes(totalSize)} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={t('physicalFiles.referenceCount')} value={totalUploads} />
          </Card>
        </Col>
      </Row>

      <Card title={t('physicalFiles.physicalFile')} style={{ marginBottom: 24 }}>
        <ProDescriptions column={2} bordered size="small">
          <ProDescriptions.Item label="ID">{physicalFile.id}</ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.slug')}><CopyCell value={physicalFile.slug} /></ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.fileHash')} span={2}>
            <CopyCell value={physicalFile.fileHash}><code style={{ fontSize: 12 }}>{physicalFile.fileHash}</code></CopyCell>
          </ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.fileSize')}>{formatBytes(physicalFile.fileSize)}</ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.mimeType')}>{physicalFile.mimeType}</ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.storagePath')} span={2}>
            <code style={{ fontSize: 12 }}>{physicalFile.storagePath}</code>
          </ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.fullPath')} span={2}>
            <code style={{ fontSize: 12 }}>{data.fullPath}</code>
          </ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.hashAlgo')}>{physicalFile.hashAlgo}</ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.status')}>
            <Tag color={STATUS_COLORS[physicalFile.status] || 'default'}>
              {physicalFile.status}
            </Tag>
          </ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.createdAt')}>{formatDate(physicalFile.createdAt)}</ProDescriptions.Item>
          <ProDescriptions.Item label={t('physicalFiles.referenceCount')}>
            {physicalFile.userFileCount + physicalFile.mediaItemCount}
          </ProDescriptions.Item>
        </ProDescriptions>
      </Card>

      {userFiles.length > 0 && (
        <>
          <h3 style={{ marginBottom: 12 }}>{t('physicalFiles.userUploadRecords')}</h3>
          <Table<CleanupQueryUserFile>
            rowKey="id"
            columns={userFileColumns}
            dataSource={userFiles}
            pagination={false}
            size="small"
          />
        </>
      )}
    </PageContainer>
  );
}