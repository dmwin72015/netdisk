import { useQuery } from '@tanstack/react-query';
import { PageContainer, ProTable } from '@ant-design/pro-components';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { get } from '@/utils/request';

interface ActivityLogRecord {
  id: string;
  userId: string;
  action: string;
  resource: string;
  ip: string;
  createdAt: string;
}

const ActivityLogsManagement = () => {
  const { t } = useTranslation();
  const { data, isLoading } = useQuery({
    queryKey: ['activityLogs'],
    queryFn: async () => {
      const result = (await get<ActivityLogRecord[]>('/admin/activity-logs')) as ActivityLogRecord[];
      return result;
    },
  });

  const columns = [
    { title: 'User ID', dataIndex: 'userId', key: 'userId' },
    { title: 'Action', dataIndex: 'action', key: 'action' },
    { title: 'Resource', dataIndex: 'resource', key: 'resource' },
    { title: 'IP', dataIndex: 'ip', key: 'ip' },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (val: unknown) => dayjs(val as string).format('YYYY-MM-DD HH:mm'),
    },
  ];

  return (
    <PageContainer>
      <ProTable<ActivityLogRecord>
        headerTitle={t('pages.activityLogs.title', 'Activity Logs')}
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

export default ActivityLogsManagement;
