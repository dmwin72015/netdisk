import { useState } from 'react';
import { LoginForm, ProFormText, ProFormCheckbox } from '@ant-design/pro-form';
import { UserOutlined, LockOutlined, GithubOutlined } from '@ant-design/icons';
import { Alert, Divider, message } from 'antd';
import { useNavigate } from 'react-router';
import { login } from '@/api/auth';

export default function LoginPage() {
  const navigate = useNavigate();
  const [errorMsg, setErrorMsg] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (values: { username: string; password: string; autoLogin?: boolean }) => {
    setLoading(true);
    setErrorMsg('');

    try {
      const data = await login(values.username, values.password);

      if (data.user.role !== 'admin') {
        setErrorMsg('仅管理员可登录后台');
        return;
      }

      localStorage.setItem('admin.token', data.tokens.accessToken);
      localStorage.setItem('admin.user', JSON.stringify(data.user));
      message.success('登录成功！');
      navigate('/dashboard');
    } catch (e) {
      setErrorMsg(e instanceof Error ? e.message : '登录失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="flex flex-col min-h-screen"
      style={{ background: 'linear-gradient(135deg, #e8f0fe 0%, #f0e6ff 50%, #e6f4ff 100%)' }}
    >
      <div className="flex-1 flex justify-center items-center py-8">
        <LoginForm
          title="Admin Panel V3"
          subTitle="网盘管理后台"
          logo={
            <svg viewBox="0 0 24 24" width="44" height="44" fill="#1890ff">
              <path d="M12 2L2 7v10l10 5 10-5V7L12 2zm0 2.18L19.18 7.5 12 10.82 4.82 7.5 12 4.18z" />
            </svg>
          }
          contentStyle={{ minWidth: 280, maxWidth: 400 }}
          initialValues={{ autoLogin: true }}
          message={
            errorMsg ? (
              <Alert message={errorMsg} type="error" showIcon className="mb-6" />
            ) : (
              false
            )
          }
          onFinish={handleSubmit}
          submitter={{
            searchConfig: { submitText: '登录' },
            submitButtonProps: { loading, size: 'large', style: { width: '100%' } },
          }}
        >
          <ProFormText
            name="username"
            fieldProps={{
              size: 'large',
              prefix: <UserOutlined />,
            }}
            placeholder="请输入用户名"
            rules={[{ required: true, message: '请输入用户名！' }]}
          />
          <ProFormText.Password
            name="password"
            fieldProps={{
              size: 'large',
              prefix: <LockOutlined />,
            }}
            placeholder="请输入密码"
            rules={[{ required: true, message: '请输入密码！' }]}
          />
          <div className="mb-6">
            <ProFormCheckbox noStyle name="autoLogin">
              自动登录
            </ProFormCheckbox>
          </div>
        </LoginForm>
      </div>
      <div className="text-center py-6 text-sm leading-[22px] text-black/45">
        <div className="mb-2">
          Admin Panel V3 &copy; {new Date().getFullYear()}
        </div>
        <div className="flex justify-center items-center gap-1">
          <span>
            <span className="text-black/25">React </span>
            <a
              href="https://react.dev"
              target="_blank"
              rel="noopener noreferrer"
              className="text-black/45 no-underline transition-colors hover:text-black/65"
            >
              19.2.7
            </a>
          </span>
          <Divider type="vertical" />
          <span>
            <span className="text-black/25">antd </span>
            <a
              href="https://ant.design"
              target="_blank"
              rel="noopener noreferrer"
              className="text-black/45 no-underline transition-colors hover:text-black/65"
            >
              6.5.0
            </a>
          </span>
          <Divider type="vertical" />
          <span>
            <span className="text-black/25">Vite </span>
            <a
              href="https://vite.dev"
              target="_blank"
              rel="noopener noreferrer"
              className="text-black/45 no-underline transition-colors hover:text-black/65"
            >
              8.1.3
            </a>
          </span>
          <Divider type="vertical" />
          <a
            href="https://github.com"
            target="_blank"
            rel="noopener noreferrer"
            className="text-black/45 no-underline transition-colors hover:text-black/65 inline-flex items-center"
          >
            <GithubOutlined className="mr-1" />
            GitHub
          </a>
        </div>
      </div>
    </div>
  );
}
