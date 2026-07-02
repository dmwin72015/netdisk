import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Form, Input, message, Modal } from 'antd';
import { PageContainer, ProTable } from '@ant-design/pro-components';
import { useState } from 'react';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { get, post } from '@/utils/request';

interface CleanupRecord {
  id: string;
  type: string;
  description: string;
  createdAt: string;
}

const CleanupManagement = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();

  const { data, isLoading } = useQuery({
    queryKey: ['cleanupRecords'],
    queryFn: async () => {
      const result = (await get<CleanupRecord[]>('/admin/cleanup/query')) as CleanupRecord[];
      return result;
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: { description: string }) => {
      return (await post('/admin/cleanup/query', values)) as undefined;
    },
    onSuccess: () => {
      message.success('Cleanup query submitted');
      setOpen(false);
      queryClient.invalidateQueries({ queryKey: ['cleanupRecords'] });
    },
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    { title: 'Description', dataIndex: 'description', key: 'description' },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (val: unknown) => dayjs(val as string).format('YYYY-MM-DD HH:mm'),
    },
  ];

  return (
    <PageContainer>
      <ProTable<CleanupRecord>
        headerTitle={t('pages.cleanup.title', 'Cleanup Management')}
        rowKey="id"
        search={{ labelWidth: 120 }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setOpen(true)}>
            Create
          </Button>,
        ]}
        columns={columns}
        dataSource={data || []}
        loading={isLoading}
      />
      <Modal
        title="Create Cleanup"
        open={open}
        onCancel={() => setOpen(false)}
        footer={null}
      >
        <Form form={form} onFinish={(values) => mutation.mutate(values as { description: string })}>
          <Form.Item
            name="description"
            label="Description"
            rules={[{ required: true, message: 'Please enter description' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={mutation.isPending}>
              Submit
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default CleanupManagement;
