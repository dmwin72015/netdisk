import { useQuery } from '@tanstack/react-query';
import { Card, Col, Row, Statistic } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { useTranslation } from 'react-i18next';
import { get } from '@/utils/request';

interface DashboardStats {
  totalUsers: number;
  totalFiles: number;
  totalStorage: number;
  todayUploads: number;
}

const Workplace = () => {
  const { t } = useTranslation();
  const { data, isLoading } = useQuery({
    queryKey: ['dashboardStats'],
    queryFn: async () => {
      const result = (await get<DashboardStats>('/admin/dashboard/stats')) as DashboardStats;
      return result;
    },
  });

  const stats = data || ({} as DashboardStats);

  return (
    <PageContainer title={t('pages.dashboard.title', 'Dashboard')}>
      <Row gutter={16}>
        <Col span={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Total Users"
              value={stats.totalUsers}
              precision={0}
              valueStyle={{ color: '#3f8600' }}
              prefix={<ArrowUpOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Total Files"
              value={stats.totalFiles}
              precision={0}
              valueStyle={{ color: '#cf1322' }}
              prefix={<ArrowDownOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Storage Used"
              value={stats.totalStorage}
              precision={2}
              suffix="GB"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={isLoading}>
            <Statistic
              title="Today Uploads"
              value={stats.todayUploads}
              precision={0}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Workplace;
