import { useQuery } from '@tanstack/react-query';
import { PageContainer, ProTable } from '@ant-design/pro-components';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { get } from '@/utils/request';

interface FileRecord {
  id: string;
  name: string;
  type: string;
  size: number;
  owner: string;
  createdAt: string;
}

const FilesManagement = () => {
  const { t } = useTranslation();
  const { data, isLoading } = useQuery({
    queryKey: ['files'],
    queryFn: async () => {
      const result = (await get<FileRecord[]>('/admin/files')) as FileRecord[];
      return result;
    },
  });

  const columns = [
    { title: 'File Name', dataIndex: 'name', key: 'name' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    {
      title: 'Size',
      dataIndex: 'size',
      key: 'size',
      render: (val: unknown) => `${((val as number) / 1024 / 1024).toFixed(2)} MB`,
    },
    { title: 'Owner', dataIndex: 'owner', key: 'owner' },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (val: unknown) => dayjs(val as string).format('YYYY-MM-DD HH:mm'),
    },
  ];

  return (
    <PageContainer>
      <ProTable<FileRecord>
        headerTitle={t('pages.files.title', 'Files Management')}
        rowKey="id"
        search={{ labelWidth: 120 }}
        toolBarRender={() => []}
        columns={columns}
        dataSource={data || []}
        loading={isLoading}
      />
    </PageContainer>
  );
};

export default FilesManagement;
