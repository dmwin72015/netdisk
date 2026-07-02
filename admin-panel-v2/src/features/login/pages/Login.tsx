import { useState } from 'react';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { LoginForm, ProFormCheckbox, ProFormText } from '@ant-design/pro-components';
import { Alert, App, Tabs } from 'antd';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Footer } from '@/components';

const Login = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [loginState, setLoginState] = useState<{ status?: string; type?: string }>({});
  const { message } = App.useApp();

  const handleSubmit = async (values: { username: string; password: string; autoLogin?: boolean }) => {
    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || t('pages.login.failure'));
      }

      const data = (await res.json()) as { data: { tokens: { accessToken: string }; user: { role: string } } };
      const token = data.data.tokens.accessToken;
      const role = data.data.user.role;

      if (values.autoLogin) {
        localStorage.setItem('nd.access', token);
      } else {
        sessionStorage.setItem('nd.access', token);
      }

      message.success(t('pages.login.success'));

      if (role === 'admin') {
        navigate('/admin', { replace: true });
      } else {
        message.error(t('pages.login.adminRequired'));
      }
    } catch (e) {
      const errorMessage = e instanceof Error ? e.message : t('pages.login.failure');
      setLoginState({ status: 'error', type: 'account' });
      message.error(errorMessage);
    }
  };

  const { status, type: loginType } = loginState;

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'auto',
        backgroundImage: "url('https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/V-_oS6r-i7wAAAAAAAAAAAAAFl94AQBr')",
        backgroundSize: '100% 100%',
      }}
    >
      <div style={{ flex: 1, padding: '32px 0', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <LoginForm
          contentStyle={{ minWidth: 280, maxWidth: '75vw' }}
          logo={<img alt="logo" src="/logo.svg" />}
          title="NetDisk Admin"
          subTitle={t('pages.layouts.userLayout.title')}
          initialValues={{ autoLogin: true }}
          onFinish={handleSubmit}
        >
          <Tabs
            activeKey="account"
            centered
            items={[
              {
                key: 'account',
                label: t('pages.login.accountLogin.tab'),
              },
              {
                key: 'mobile',
                label: t('pages.login.phoneLogin.tab'),
                disabled: true,
              },
            ]}
          />

          {status === 'error' && loginType === 'account' && (
            <Alert
              style={{ marginBottom: 24 }}
              message={t('pages.login.accountLogin.errorMessage')}
              type="error"
              showIcon
            />
          )}

          <ProFormText
            name="username"
            fieldProps={{
              size: 'large',
              prefix: <UserOutlined />,
            }}
            placeholder={t('pages.login.username.placeholder')}
            rules={[
              {
                required: true,
                message: t('pages.login.username.required'),
              },
            ]}
          />
          <ProFormText.Password
            name="password"
            fieldProps={{
              size: 'large',
              prefix: <LockOutlined />,
            }}
            placeholder={t('pages.login.password.placeholder')}
            rules={[
              {
                required: true,
                message: t('pages.login.password.required'),
              },
            ]}
          />

          <div style={{ marginBottom: 24 }}>
            <ProFormCheckbox noStyle name="autoLogin">
              {t('pages.login.rememberMe')}
            </ProFormCheckbox>
          </div>
        </LoginForm>
      </div>
      <Footer />
    </div>
  );
};

export default Login;
