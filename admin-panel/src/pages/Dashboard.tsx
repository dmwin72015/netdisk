import { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Spin } from 'antd';
import {
  UserOutlined,
  FileOutlined,
  DatabaseOutlined,
  FireOutlined,
} from '@ant-design/icons';
import { adminGetDashboard, adminListStorageStats, type AdminDashboard, type AdminStorageStat } from '../api/admin';

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function Dashboard() {
  const [dashboard, setDashboard] = useState<AdminDashboard | null>(null);
  const [storage, setStorage] = useState<AdminStorageStat[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [d, s] = await Promise.all([
          adminGetDashboard(),
          adminListStorageStats(20),
        ]);
        setDashboard(d);
        setStorage(s.items);
      } catch (e) {
        console.error(e);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 60 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!dashboard) {
    return <div>Failed to load dashboard data.</div>;
  }

  const maxStorage = Math.max(...storage.map(s => s.usedBytes), 1);

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>Dashboard</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Users"
              value={dashboard.totalUsers}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Files"
              value={dashboard.totalFiles}
              prefix={<FileOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Storage"
              value={dashboard.totalStorage}
              formatter={(v) => formatBytes(v as number)}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Today Active"
              value={dashboard.todayActive}
              prefix={<FireOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Top Storage Users" style={{ marginTop: 24 }}>
        {storage.length === 0 && (
          <div style={{ color: '#999', textAlign: 'center', padding: 20 }}>
            No data
          </div>
        )}
        {storage.map((s) => {
          const pct = (s.usedBytes / maxStorage) * 100;
          const usagePct = s.totalBytes > 0
            ? Math.round((s.usedBytes / s.totalBytes) * 100)
            : 0;
          return (
            <div key={s.id} style={{ marginBottom: 16 }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  marginBottom: 4,
                }}
              >
                <span>{s.username}</span>
                <span style={{ color: '#666', fontSize: 12 }}>
                  {formatBytes(s.usedBytes)} / {formatBytes(s.totalBytes)} ({usagePct}%)
                </span>
              </div>
              <div
                style={{
                  height: 20,
                  background: '#f5f5f5',
                  borderRadius: 4,
                  overflow: 'hidden',
                }}
              >
                <div
                  style={{
                    height: '100%',
                    width: `${pct}%`,
                    background: pct > 80 ? '#ff4d4f' : '#1890ff',
                    borderRadius: 4,
                    transition: 'width 0.3s',
                    minWidth: usagePct > 0 ? 2 : 0,
                  }}
                />
              </div>
            </div>
          );
        })}
      </Card>
    </div>
  );
}
