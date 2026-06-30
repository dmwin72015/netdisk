import { Card, Row, Col, Statistic, Spin, Result, Progress } from 'antd';
import {
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  UserAddOutlined,
  FileAddOutlined,
  HddOutlined,
} from '@ant-design/icons';
import { useDashboardStats } from '../api/admin.hooks';

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function Dashboard() {
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
      <div style={{ padding: 24 }}>
        <Result
          status="error"
          title="Failed to load dashboard"
          subTitle={error.message}
        />
      </div>
    );
  }

  if (!stats) {
    return (
      <div style={{ padding: 24 }}>
        <Result status="warning" title="No dashboard data" />
      </div>
    );
  }

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>Dashboard</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Total Users"
              value={stats.totalUsers}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Total Files"
              value={stats.totalFiles}
              prefix={<FileOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Storage Used"
              value={stats.storageUsed}
              formatter={(v) => formatBytes(v as number)}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Total Quota"
              value={stats.totalStorage}
              formatter={(v) => formatBytes(v as number)}
              prefix={<HddOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="New Users Today"
              value={stats.newTodayUsers}
              prefix={<UserAddOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="New Files Today"
              value={stats.newTodayFiles}
              prefix={<FileAddOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="Storage Usage">
            <Progress
              percent={Math.round((stats.storageUsed / stats.totalStorage) * 100)}
              status={stats.storageUsed >= stats.totalStorage ? 'exception' : 'active'}
            />
            <div style={{ marginTop: 8, color: '#888' }}>
              {formatBytes(stats.storageUsed)} used of {formatBytes(stats.totalStorage)}
            </div>
          </Card>
        </Col>
        {stats.diskTotal > 0 && (
          <Col xs={24} lg={12}>
            <Card title="System Disk">
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
                {formatBytes(stats.diskFree)} free of {formatBytes(stats.diskTotal)}
              </div>
            </Card>
          </Col>
        )}
      </Row>
    </div>
  );
}