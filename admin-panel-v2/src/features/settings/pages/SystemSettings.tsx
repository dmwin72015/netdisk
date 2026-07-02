import { useQuery } from '@tanstack/react-query';
import { Card, Form, Input, Switch, Button, message, Space } from 'antd';
import { PageContainer } from '@ant-design/pro-components';
import { SaveOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { get, post } from '@/utils/request';

interface SystemConfig {
  siteName: string;
  maintenanceMode: boolean;
  maxUploadSize: number;
  allowedFileTypes: string;
}

const SystemSettings = () => {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['systemConfig'],
    queryFn: async () => {
      const result = (await get<SystemConfig>('/admin/system-config')) as SystemConfig;
      return result;
    },
  });

  const [form] = Form.useForm();

  const onFinish = async (values: SystemConfig) => {
    try {
      await post('/admin/system-config', values);
      message.success('Settings saved');
      refetch();
    } catch {
      message.error('Failed to save settings');
    }
  };

  if (isLoading || !data) {
    return <PageContainer loading>{t('common.loading', 'Loading...')}</PageContainer>;
  }

  return (
    <PageContainer title={t('pages.settings.title', 'System Settings')}>
      <Card>
        <Form
          form={form}
          layout="vertical"
          initialValues={data}
          onFinish={onFinish}
        >
          <Form.Item name="siteName" label="Site Name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="maintenanceMode" label="Maintenance Mode" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="maxUploadSize" label="Max Upload Size (MB)" rules={[{ required: true }]}>
            <Input type="number" />
          </Form.Item>
          <Form.Item name="allowedFileTypes" label="Allowed File Types">
            <Input />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />}>
                Save
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
};

export default SystemSettings;
