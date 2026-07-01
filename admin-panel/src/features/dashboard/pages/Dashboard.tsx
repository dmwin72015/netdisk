import { Card, Row, Col, Statistic, Spin, Result, Progress } from 'antd';
import {
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  UserAddOutlined,
  FileAddOutlined,
  HddOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useDashboardStats } from '../../../api/admin.hooks';
import { PageContainer } from '../../../components/PageContainer';
import { formatBytes } from '../../../utils/format';

export default function DashboardPage() {
  const { t } = useTranslation();
  const { data: stats, isLoading, error } = useDashboardStats();

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
          title={t('dashboard.failed')}
          subTitle={error.message}
        />
      </PageContainer>
    );
  }

  if (!stats) {
    return (
      <PageContainer>
        <Result status="warning" title={t('dashboard.noData')} />
      </PageContainer>
    );
  }

  return (
    <PageContainer title={t('dashboard.title')}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title={t('dashboard.totalUsers')}
              value={stats.totalUsers}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title={t('dashboard.totalFiles')}
              value={stats.totalFiles}
              prefix={<FileOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title={t('dashboard.storageUsed')}
              value={stats.storageUsed}
              formatter={(v) => formatBytes(v as number)}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title={t('dashboard.totalQuota')}
              value={stats.totalStorage}
              formatter={(v) => formatBytes(v as number)}
              prefix={<HddOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title={t('dashboard.newUsersToday')}
              value={stats.newTodayUsers}
              prefix={<UserAddOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title={t('dashboard.newFilesToday')}
              value={stats.newFilesToday}
              prefix={<FileAddOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title={t('dashboard.storageUsage')}>
            <Progress
              percent={Math.round((stats.storageUsed / stats.totalStorage) * 100)}
              status={stats.storageUsed >= stats.totalStorage ? 'exception' : 'active'}
            />
            <div style={{ marginTop: 8, color: '#888' }}>
              {t('dashboard.usedOf', { used: formatBytes(stats.storageUsed), total: formatBytes(stats.totalStorage) })}
            </div>
          </Card>
        </Col>
        {stats.diskTotal > 0 && (
          <Col xs={24} lg={12}>
            <Card title={t('dashboard.systemDisk')}>
              <Progress
                percent={Math.round((stats.diskUsed / stats.diskTotal) * 100)}
                strokeColor={
                  stats.diskUsed / stats.diskTotal < 0.7
                    ? '#52c41a'
                    : stats.diskUsed / stats.diskTotal < 0.9
                      ? '#faad14'
                      : '#ff4d4f'
                }
              />
              <div style={{ marginTop: 8, color: '#888' }}>
                {t('dashboard.freeOf', { free: formatBytes(stats.diskFree), total: formatBytes(stats.diskTotal) })}
              </div>
            </Card>
          </Col>
        )}
      </Row>
    </PageContainer>
  );
}