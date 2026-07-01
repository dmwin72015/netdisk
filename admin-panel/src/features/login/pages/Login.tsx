import { useState } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

interface LoginResponse {
  user: {
    role: string;
    [key: string]: unknown;
  };
  tokens: {
    accessToken: string;
  };
}

export default function LoginPage() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true);
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || t('login.failed'));
      }

      const data = (await res.json()) as { data: LoginResponse };
      localStorage.setItem('nd.access', data.data.tokens.accessToken);
      localStorage.setItem('nd.user', JSON.stringify(data.data.user));

      if (data.data.user.role === 'admin') {
        navigate('/admin');
      } else {
        message.error(t('login.adminRequired'));
      }
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('login.failed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: '#f0f2f5',
      }}
    >
      <Card title={t('login.title')} style={{ width: 400 }}>
        <Form name="login" onFinish={onFinish} autoComplete="off" size="large">
          <Form.Item
            name="email"
            rules={[
              { required: true, message: t('login.emailPlaceholder') },
              { type: 'email', message: t('login.emailError') },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder={t('login.emailPlaceholder')} />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[
              { required: true, message: t('login.passwordPlaceholder') },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t('login.passwordPlaceholder')} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              {loading ? t('login.loggingIn') : t('login.login')}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
