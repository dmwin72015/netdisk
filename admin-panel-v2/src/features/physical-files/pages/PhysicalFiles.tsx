import { useQuery } from '@tanstack/react-query';
import { ProTable } from '@ant-design/pro-components';
import { Tag } from 'antd';
import dayjs from 'dayjs';
import { get } from '@/utils/request';

interface PhysicalFileRecord {
  id: string;
  path: string;
  size: number;
  checksum: string;
  status: string;
  createdAt: string;
}

const PhysicalFiles = () => {
  const { data, isLoading } = useQuery({
    queryKey: ['physicalFiles'],
    queryFn: async () => {
      const result = (await get<PhysicalFileRecord[]>('/admin/physical-files')) as PhysicalFileRecord[];
      return result;
    },
  });

  const columns = [
    { title: 'Path', dataIndex: 'path', key: 'path' },
    {
      title: 'Size',
      dataIndex: 'size',
      key: 'size',
      render: (val: unknown) => `${((val as number) / 1024 / 1024).toFixed(2)} MB`,
    },
    { title: 'Checksum', dataIndex: 'checksum', key: 'checksum' },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (val: unknown) => {
        const status = val as string;
        return <Tag color={status === 'active' ? 'green' : 'red'}>{status}</Tag>;
      },
    },
    {
      title: 'Created At',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (val: unknown) => dayjs(val as string).format('YYYY-MM-DD HH:mm'),
    },
  ];

  return (
    <div className="p-6">
      <ProTable<PhysicalFileRecord>
        headerTitle="Physical Files"
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        columns={columns}
        dataSource={data || []}
        loading={isLoading}
      />
    </div>
  );
};

export default PhysicalFiles;
