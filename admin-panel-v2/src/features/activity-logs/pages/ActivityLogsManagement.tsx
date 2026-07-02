import { useQuery } from '@tanstack/react-query';
import { ProTable } from '@ant-design/pro-components';
import dayjs from 'dayjs';
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
    <div className="p-6">
      <ProTable<ActivityLogRecord>
        headerTitle="Activity Logs"
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        columns={columns}
        dataSource={data || []}
        loading={isLoading}
      />
    </div>
  );
};

export default ActivityLogsManagement;
