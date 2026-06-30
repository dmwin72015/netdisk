import { useEffect, useState } from 'react';
import { Form, Input, InputNumber, Switch, Button, message, Card, Typography, Alert } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { adminGetSettings, adminUpdateSettings, type AdminSettings } from '../api/admin';

const { Title } = Typography;

function parseBytesInput(val: number): number {
  return Math.round(val * 1024 * 1024 * 1024);
}

export default function Settings() {
  const [settings, setSettings] = useState<AdminSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    adminGetSettings()
      .then((s) => {
        setSettings(s);
        form.setFieldsValue({
          siteName: s.siteName,
          allowRegistration: s.allowRegistration,
          defaultQuota: parseFloat((s.defaultQuota / 1073741824).toFixed(2)),
          maxUploadSize: parseFloat((s.maxUploadSize / 1073741824).toFixed(2)),
        });
      })
      .catch(() => message.error('Failed to load settings'))
      .finally(() => setLoading(false));
  }, [form]);

  const onFinish = async (values: {
    siteName: string;
    allowRegistration: boolean;
    defaultQuota: number;
    maxUploadSize: number;
  }) => {
    setSaving(true);
    try {
      const updated = await adminUpdateSettings({
        siteName: values.siteName,
        allowRegistration: values.allowRegistration,
        defaultQuota: parseBytesInput(values.defaultQuota),
        maxUploadSize: parseBytesInput(values.maxUploadSize),
      });
      setSettings(updated);
      message.success('Settings saved successfully');
    } catch {
      message.error('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 60 }}>
        <div>Loading settings...</div>
      </div>
    );
  }

  return (
    <div>
      <Title level={4}>System Settings</Title>
      <Alert
        message="Changes take effect immediately"
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />
      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            siteName: settings?.siteName || '',
            allowRegistration: settings?.allowRegistration ?? true,
            defaultQuota: settings?.defaultQuota ? parseFloat((settings.defaultQuota / 1073741824).toFixed(2)) : 10,
            maxUploadSize: settings?.maxUploadSize ? parseFloat((settings.maxUploadSize / 1073741824).toFixed(2)) : 100,
          }}
        >
          <Form.Item
            name="siteName"
            label="Site Name"
            rules={[{ required: true, message: 'Please input site name' }]}
          >
            <Input placeholder="NetDisk" />
          </Form.Item>
          <Form.Item
            name="allowRegistration"
            label="Allow Registration"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
          <Form.Item
            name="defaultQuota"
            label="Default Quota (GB)"
            rules={[{ required: true, message: 'Please input default quota' }]}
          >
            <InputNumber min={0} precision={2} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            name="maxUploadSize"
            label="Max Upload Size (GB)"
            rules={[{ required: true, message: 'Please input max upload size' }]}
          >
            <InputNumber min={0} precision={2} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving} block>
              Save Settings
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
