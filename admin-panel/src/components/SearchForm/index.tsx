import type { ReactNode } from 'react';
import { Form, Row, Col, Button, Space } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

export interface SearchField {
  name: string;
  label: string;
  children: ReactNode;
  span?: number;
}

interface SearchFormProps {
  fields: SearchField[];
  onSearch: (values: Record<string, unknown>) => void;
  onReset: () => void;
  loading?: boolean;
  initialValues?: Record<string, unknown>;
}

export default function SearchForm({
  fields,
  onSearch,
  onReset,
  loading,
  initialValues,
}: SearchFormProps) {
  const { t } = useTranslation();
  const [form] = Form.useForm();

  const handleReset = () => {
    form.resetFields();
    onReset();
  };

  return (
    <Form
      form={form}
      layout="inline"
      initialValues={initialValues}
      onFinish={onSearch}
      style={{ marginBottom: 16, flexWrap: 'wrap', gap: 8 }}
    >
      <Row gutter={[16, 12]} style={{ width: '100%' }}>
        {fields.map((field) => (
          <Col key={field.name} xs={24} sm={12} md={field.span || 6}>
            <Form.Item name={field.name} label={field.label} style={{ width: '100%' }}>
              {field.children}
            </Form.Item>
          </Col>
        ))}
        <Col>
          <Space>
            <Button
              type="primary"
              htmlType="submit"
              icon={<SearchOutlined />}
              loading={loading}
            >
              {t('common.search')}
            </Button>
            <Button onClick={handleReset} icon={<ReloadOutlined />}>
              {t('common.reset') || '重置'}
            </Button>
          </Space>
        </Col>
      </Row>
    </Form>
  );
}
