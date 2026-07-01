import { Spin, Card, Row, Col, Tag, Avatar, Button, Result } from 'antd';
import { ProDescriptions } from '@ant-design/pro-components';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { useUser } from '../../../api/admin.hooks';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: user, isLoading, error } = useUser(id!);

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
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/admin/users')}
          style={{ marginBottom: 16 }}
        >
          {t('common.back')}
        </Button>
        <Result
          status="error"
          title={t('userDetail.failed')}
          subTitle={error.message}
        />
      </div>
    );
  }

  if (!user) {
    return (
      <div style={{ padding: 24 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate('/admin/users')}
          style={{ marginBottom: 16 }}
        >
          {t('common.back')}
        </Button>
        <div style={{ color: '#999', textAlign: 'center', padding: 40 }}>{t('userDetail.notFound')}</div>
      </div>
    );
  }

  const usagePct = user.totalBytes > 0 ? Math.round((user.usedBytes / user.totalBytes) * 100) : 0;

  return (
    <div style={{ padding: 24 }}>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/admin/users')}
        style={{ marginBottom: 16 }}
      >
        {t('userDetail.back')}
      </Button>
      <h2 style={{ marginBottom: 24 }}>{t('userDetail.title')}: {user.username}</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title={t('userDetail.basicInfo')}>
            <ProDescriptions column={1} size="small" style={{ margin: 0 }}>
              <ProDescriptions.Item label={t('users.username')}>{user.username}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.slug')}>{user.slug}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('users.email')}>{user.email}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('users.role')}>
                <Tag color={ROLE_COLORS[user.role] || 'default'}>{user.role}</Tag>
              </ProDescriptions.Item>
              <ProDescriptions.Item label={t('users.registerMethod')}>{user.registerMethod}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.status')}>{user.status}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('users.registered')}>{formatDate(user.createdAt)}</ProDescriptions.Item>
            </ProDescriptions>
            {user.profile && (
              <div style={{ marginTop: 16 }}>
                <strong>{t('userDetail.profile')}</strong>
                <ProDescriptions column={1} size="small" style={{ marginTop: 8 }}>
                  {user.profile.displayName && (
                    <ProDescriptions.Item label={t('userDetail.displayName')}>{user.profile.displayName}</ProDescriptions.Item>
                  )}
                  {user.profile.avatarUrl && (
                    <ProDescriptions.Item label={t('userDetail.avatar')}>
                      <Avatar src={user.profile.avatarUrl} />
                    </ProDescriptions.Item>
                  )}
                  {user.profile.bio && (
                    <ProDescriptions.Item label={t('userDetail.bio')}>{user.profile.bio}</ProDescriptions.Item>
                  )}
                </ProDescriptions>
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title={t('userDetail.storage')}>
            <ProDescriptions column={1} size="small">
              <ProDescriptions.Item label={t('userDetail.used')}>{formatBytes(user.usedBytes)}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.base')}>{formatBytes(user.baseBytes)}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.memberBonus')}>{formatBytes(user.memberBonusBytes)}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.pack')}>{formatBytes(user.packBytes)}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.total')}>{formatBytes(user.totalBytes)}</ProDescriptions.Item>
              <ProDescriptions.Item label={t('userDetail.usage')}>
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
              </ProDescriptions.Item>
            </ProDescriptions>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title={t('userDetail.oauth')}>
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
              <span style={{ color: '#999' }}>{t('userDetail.noOauth')}</span>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
}