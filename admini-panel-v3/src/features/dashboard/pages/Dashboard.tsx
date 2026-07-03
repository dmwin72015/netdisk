import { Card, Row, Col, Statistic, Spin, Result, Progress } from 'antd';
import {
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  UserAddOutlined,
  FileAddOutlined,
  HddOutlined,
} from '@ant-design/icons';
import { useDashboardStats } from '@/api/admin.hooks';
import PageContainer from '@/components/PageContainer';
import { formatBytes } from '@/utils/format';

export default function DashboardPage() {
  const { data: stats, isLoading, error } = useDashboardStats();

  if (isLoading) {
    return (
      <div className="text-center p-15">
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return (
      <PageContainer>
        <Result status="error" title="加载失败" subTitle={error.message} />
      </PageContainer>
    );
  }

  if (!stats) {
    return (
      <PageContainer>
        <Result status="warning" title="暂无数据" />
      </PageContainer>
    );
  }

  return (
    <PageContainer title="仪表盘">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable>
            <Statistic title="总用户数" value={stats.totalUsers} prefix={<UserOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable>
            <Statistic title="总文件数" value={stats.totalFiles} prefix={<FileOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable>
            <Statistic title="已用存储" value={stats.storageUsed} formatter={(v) => formatBytes(v as number)} prefix={<DatabaseOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable>
            <Statistic title="总配额" value={stats.totalStorage} formatter={(v) => formatBytes(v as number)} prefix={<HddOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable>
            <Statistic title="今日新用户" value={stats.newTodayUsers} prefix={<UserAddOutlined />} valueStyle={{ color: '#52c41a' }} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card hoverable>
            <Statistic title="今日新文件" value={stats.newTodayFiles} prefix={<FileAddOutlined />} valueStyle={{ color: '#1890ff' }} />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} className="mt-6">
        <Col xs={24} lg={12}>
          <Card title="存储使用率" hoverable>
            <Progress
              percent={Math.round((stats.storageUsed / stats.totalStorage) * 100)}
              status={stats.storageUsed >= stats.totalStorage ? 'exception' : 'active'}
            />
            <div className="mt-2 text-[#888]">
              已用 {formatBytes(stats.storageUsed)} / 共 {formatBytes(stats.totalStorage)}
            </div>
          </Card>
        </Col>
        {stats.diskTotal > 0 && (
          <Col xs={24} lg={12}>
            <Card title="系统磁盘" hoverable>
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
              <div className="mt-2 text-[#888]">
                可用 {formatBytes(stats.diskFree)} / 共 {formatBytes(stats.diskTotal)}
              </div>
            </Card>
          </Col>
        )}
      </Row>
    </PageContainer>
  );
}
