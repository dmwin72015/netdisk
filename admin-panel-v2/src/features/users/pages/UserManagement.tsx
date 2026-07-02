import { useQuery } from '@tanstack/react-query';
import { ProTable } from '@ant-design/pro-components';
import { Tag } from 'antd';
import dayjs from 'dayjs';
import { get } from '@/utils/request';

interface UserRecord {
  id: string;
  email: string;
  name: string;
  slug: string;
  registerMethod: string;
  createdAt: string;
  status: string;
}

const UserManagement = () => {
  const { data, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const result = (await get<UserRecord[]>('/admin/users')) as UserRecord[];
      return result;
    },
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id' },
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Email', dataIndex: 'email', key: 'email' },
    { title: 'Slug', dataIndex: 'slug', key: 'slug' },
    { title: 'Register Method', dataIndex: 'registerMethod', key: 'registerMethod' },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (val: unknown) => dayjs(val as string).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (val: unknown) => {
        const status = val as string;
        return <Tag color={status === 'active' ? 'green' : 'red'}>{status}</Tag>;
      },
    },
  ];

  return (
    <div className="p-6">
      <ProTable<UserRecord>
        headerTitle="User Management"
        rowKey="id"
        search={{
          labelWidth: 'auto',
        }}
        toolBarRender={() => ['Create', 'Edit', 'Delete']}
        columns={columns}
        dataSource={data || []}
        loading={isLoading}
      />
    </div>
  );
};

export default UserManagement;
