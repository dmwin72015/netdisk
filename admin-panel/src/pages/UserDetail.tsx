import { useEffect, useState } from 'react';
import { Spin, Card, Row, Col, Descriptions, Tag, Avatar, Button, message } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { adminGetUser, type AdminUser } from '../api/admin';

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(epoch: number): string {
  return new Date(epoch * 1000).toLocaleString();
}

const ROLE_COLORS: Record<string, string> = {
  admin: 'red',
  user: 'blue',
  moderator: 'orange',
};

export default function UserDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [user, setUser] = useState<AdminUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    adminGetUser(id)
      .then(setUser)
      .catch(() => message.error('Failed to load user'))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 60 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!user) {
    return (
      <div>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/users')} style={{ marginBottom: 16 }}>
          Back
        </Button>
        <div>User not found.</div>
      </div>
    );
  }

  const usagePct = user.totalBytes > 0 ? Math.round((user.usedBytes / user.totalBytes) * 100) : 0;

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/users')} style={{ marginBottom: 16 }}>
        Back to Users
      </Button>
      <h2 style={{ marginBottom: 24 }}>User: {user.username}</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title="Basic Info">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Username">{user.username}</Descriptions.Item>
              <Descriptions.Item label="Slug">{user.slug}</Descriptions.Item>
              <Descriptions.Item label="Email">{user.email}</Descriptions.Item>
              <Descriptions.Item label="Role">
                <Tag color={ROLE_COLORS[user.role] || 'default'}>{user.role}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Register Method">{user.registerMethod}</Descriptions.Item>
              <Descriptions.Item label="Status">{user.status}</Descriptions.Item>
              <Descriptions.Item label="Created">{formatDate(user.createdAt)}</Descriptions.Item>
            </Descriptions>
            {user.profile && (
              <div style={{ marginTop: 16 }}>
                <strong>Profile</strong>
                <Descriptions column={1} size="small" style={{ marginTop: 8 }}>
                  {user.profile.displayName && (
                    <Descriptions.Item label="Display Name">{user.profile.displayName}</Descriptions.Item>
                  )}
                  {user.profile.avatarUrl && (
                    <Descriptions.Item label="Avatar">
                      <Avatar src={user.profile.avatarUrl} />
                    </Descriptions.Item>
                  )}
                  {user.profile.bio && (
                    <Descriptions.Item label="Bio">{user.profile.bio}</Descriptions.Item>
                  )}
                </Descriptions>
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="Storage">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Used">{formatBytes(user.usedBytes)}</Descriptions.Item>
              <Descriptions.Item label="Base">{formatBytes(user.baseBytes)}</Descriptions.Item>
              <Descriptions.Item label="Total">{formatBytes(user.totalBytes)}</Descriptions.Item>
              <Descriptions.Item label="Usage">
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <div style={{ flex: 1, height: 8, background: '#f5f5f5', borderRadius: 4, overflow: 'hidden' }}>
                    <div
                      style={{
                        height: '100%',
                        width: `${usagePct}%`,
                        background: usagePct > 80 ? '#ff4d4f' : '#1890ff',
                        borderRadius: 4,
                      }}
                    />
                  </div>
                  <span>{usagePct}%</span>
                </div>
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="OAuth Accounts">
            {user.oauthAccounts && user.oauthAccounts.length > 0 ? (
              user.oauthAccounts.map((acc) => (
                <div key={acc.provider} style={{ marginBottom: 8 }}>
                  <Tag color="blue">{acc.provider}</Tag>
                  <span style={{ color: '#666', fontSize: 12 }}>
                    {acc.providerAccountId} &middot; {formatDate(acc.createdAt)}
                  </span>
                </div>
              ))
            ) : (
              <span style={{ color: '#999' }}>No OAuth accounts</span>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
}
