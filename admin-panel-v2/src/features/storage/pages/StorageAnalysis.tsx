import { useQuery } from '@tanstack/react-query';
import { Card, Col, Row, Statistic } from 'antd';
import { DatabaseOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';
import { get } from '@/utils/request';

interface StorageStats {
  totalStorage: number;
  usedStorage: number;
  freeStorage: number;
  usagePercent: number;
}

const StorageAnalysis = () => {
  const { t } = useTranslation();
  const { data, isLoading } = useQuery({
    queryKey: ['storageStats'],
    queryFn: async () => {
      const result = (await get<StorageStats>('/admin/storage')) as StorageStats;
      return result;
    },
  });

  const stats = data || ({} as StorageStats);

  return (
    <PageContainer title={t('pages.storage.title', 'Storage Analysis')}>
      <Row gutter={16}>
        <Col span={8}>
          <Card loading={isLoading}>
            <Statistic
              title="Total Storage"
              value={stats.totalStorage}
              precision={2}
              suffix="GB"
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card loading={isLoading}>
            <Statistic
              title="Used Storage"
              value={stats.usedStorage}
              precision={2}
              suffix="GB"
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card loading={isLoading}>
            <Statistic
              title="Free Storage"
              value={stats.freeStorage}
              precision={2}
              suffix="GB"
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default StorageAnalysis;
