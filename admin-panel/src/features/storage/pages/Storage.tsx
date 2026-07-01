import { Card, Row, Col, Spin, Result, Progress } from 'antd';
import { useTranslation } from 'react-i18next';
import { useStorageStats } from '../../../api/admin.hooks';
import PageContainer from '../../../components/PageContainer';
import { formatBytes } from '../../../utils/format';

const CATEGORY_COLORS: Record<string, string> = {
  video: '#722ed1',
  audio: '#13c2c2',
  image: '#1890ff',
  document: '#52c41a',
  archive: '#fa8c16',
  other: '#eb2f96',
  trash: '#ff4d4f',
};

export default function StoragePage() {
  const { t } = useTranslation();
  const { data: categories, isLoading, error } = useStorageStats();

  const CATEGORY_LABELS: Record<string, string> = {
    video: t('files.video'),
    audio: t('files.audio'),
    image: t('files.image'),
    document: t('files.document'),
    archive: t('files.archive'),
    other: t('files.other'),
    trash: t('storage.trash'),
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
      <PageContainer>
        <Result
          status="error"
          title={t('storage.failed')}
          subTitle={error.message}
        />
      </PageContainer>
    );
  }

  if (!categories || categories.length === 0) {
    return (
      <PageContainer>
        <Result status="warning" title={t('storage.noData')} />
      </PageContainer>
    );
  }

  const totalBytes = categories.reduce((sum, c) => sum + c.totalSize, 0);

  const trashCategory = categories.find((c) => c.category === 'trash');
  const regularCategories = categories.filter((c) => c.category !== 'trash');

  return (
    <PageContainer title={t('storage.title')}>
      <Row gutter={[16, 16]}>
        {regularCategories.map((stat) => {
          const pct = totalBytes > 0 ? Math.round((stat.totalSize / totalBytes) * 100) : 0;
          return (
            <Col xs={24} sm={12} lg={8} key={stat.category}>
              <Card
                title={CATEGORY_LABELS[stat.category] || stat.category.charAt(0).toUpperCase() + stat.category.slice(1)}
                size="small"
              >
                <div style={{ marginBottom: 8 }}>{t('storage.files_count', { count: stat.fileCount })}</div>
                <div style={{ marginBottom: 8, fontSize: 18, fontWeight: 600, color: CATEGORY_COLORS[stat.category] || '#1890ff' }}>
                  {formatBytes(stat.totalSize)}
                </div>
                <Progress
                  percent={pct}
                  strokeColor={CATEGORY_COLORS[stat.category] || '#1890ff'}
                  size="small"
                  showInfo={false}
                />
                <div style={{ textAlign: 'right', fontSize: 12, color: '#999', marginTop: 4 }}>
                  {t('storage.pctOfTotal', { pct })}
                </div>
              </Card>
            </Col>
          );
        })}
      </Row>

      {trashCategory && (
        <Card
          title={t('storage.trash')}
          size="small"
          style={{ marginTop: 24 }}
        >
          <Row gutter={16} align="middle">
            <Col>
              <span style={{ color: CATEGORY_COLORS.trash, fontSize: 18, fontWeight: 600 }}>
                {formatBytes(trashCategory.totalSize)}
              </span>
            </Col>
            <Col>
              <span style={{ color: '#999' }}>{t('storage.files_count', { count: trashCategory.fileCount })}</span>
            </Col>
            <Col flex="auto">
              <Progress
                percent={totalBytes > 0 ? Math.round((trashCategory.totalSize / totalBytes) * 100) : 0}
                strokeColor={CATEGORY_COLORS.trash}
                size="small"
              />
            </Col>
          </Row>
        </Card>
      )}
    </PageContainer>
  );
}