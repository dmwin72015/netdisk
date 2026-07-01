import { useCallback } from 'react';
import { Space, Tag, Button, Popconfirm } from 'antd';
import { UndoOutlined, DeleteOutlined } from '@ant-design/icons';
import { Link } from 'react-router';
import type { ColumnsType } from 'antd/es/table';
import { Input, Select } from 'antd';
import { useTranslation } from 'react-i18next';
import PageContainer from '../../../components/PageContainer';
import SearchForm from '../../../components/SearchForm';
import type { SearchField } from '../../../components/SearchForm';
import ProTable from '../../../components/ProTable';
import { useTableUrlState } from '../../../hooks/useTableUrlState';
import { useFiles, useDeleteFile, useRestoreFile } from '../../../api/admin.hooks';
import type { AdminFile } from '../../../api/admin';
import { formatBytes, formatDate } from '../../../utils/format';

const CATEGORY_OPTIONS = [
  { label: '文档', value: 'document' },
  { label: '图片', value: 'image' },
  { label: '视频', value: 'video' },
  { label: '音频', value: 'audio' },
  { label: '压缩包', value: 'archive' },
  { label: '其他', value: 'other' },
];

const STATUS_OPTIONS = [
  { label: '正常', value: 'active' },
  { label: '已删除', value: 'trashed' },
];

export default function FilesPage() {
  const { t } = useTranslation();
  const deleteFileMut = useDeleteFile();
  const restoreFileMut = useRestoreFile();

  const { params, setParams, resetParams } = useTableUrlState({
    page: 1,
    pageSize: 20,
    search: '',
    fileCategory: '',
    isTrashed: '',
  });

  const queryResult = useFiles({
    limit: params.pageSize,
    offset: (params.page - 1) * params.pageSize,
    search: params.search || undefined,
    fileCategory: params.fileCategory || undefined,
    isTrashed:
      params.isTrashed === 'trashed' ? true : params.isTrashed === 'active' ? false : undefined,
  });

  const handleSearch = useCallback(
    (values: Record<string, unknown>) => {
      setParams({
        page: 1,
        search: (values.search as string) || '',
        fileCategory: (values.fileCategory as string) || '',
        isTrashed: (values.isTrashed as string) || '',
      });
    },
    [setParams],
  );

  const handlePageChange = useCallback(
    (page: number, pageSize: number) => {
      setParams({ page, pageSize });
    },
    [setParams],
  );

  const columns: ColumnsType<AdminFile> = [
    { title: t('files.filename'), dataIndex: 'fileName', ellipsis: true },
    {
      title: t('files.owner'),
      dataIndex: 'username',
      render: (_, r) => <Link to={`/admin/users/${r.userId}`}>{r.username}</Link>,
      width: 150,
    },
    {
      title: t('files.type'),
      dataIndex: 'fileCategory',
      render: (v: string) => <Tag color="blue">{v || t('files.other')}</Tag>,
      width: 100,
    },
    {
      title: t('files.size'),
      dataIndex: 'fileSize',
      render: (v: number) => formatBytes(v),
      width: 100,
    },
    {
      title: t('files.status'),
      key: 'status',
      render: (_, r) => (
        <Space>
          {r.isTrashed && <Tag color="red">{t('files.deleted')}</Tag>}
          {r.isStarred && <Tag color="gold">{t('files.starred')}</Tag>}
          {!r.isTrashed && !r.isStarred && <Tag>{t('files.normal')}</Tag>}
        </Space>
      ),
      width: 140,
    },
    {
      title: t('files.uploaded'),
      dataIndex: 'createdAt',
      render: (v: number) => formatDate(v),
      width: 160,
    },
    {
      title: t('files.actions'),
      width: 180,
      render: (_, r) => (
        <Space>
          {r.isTrashed && (
            <Popconfirm
              title={t('files.restore')}
              description={t('files.restoreConfirm')}
              onConfirm={() => restoreFileMut.mutate(r.id)}
              okText={t('common.yes')}
              cancelText={t('common.no')}
            >
              <Button type="link" size="small" icon={<UndoOutlined />}>
                {t('files.restore')}
              </Button>
            </Popconfirm>
          )}
          <Popconfirm
            title={t('files.permanentDelete')}
            description={t('files.noUndo')}
            onConfirm={() => deleteFileMut.mutate(r.id)}
            okText={t('common.yes')}
            cancelText={t('common.no')}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const searchFields: SearchField[] = [
    {
      name: 'search',
      label: t('files.filename'),
      children: <Input placeholder={t('files.search')} allowClear />,
    },
    {
      name: 'fileCategory',
      label: t('files.category'),
      children: (
        <Select
          allowClear
          placeholder={t('files.allCategories')}
          options={CATEGORY_OPTIONS}
          style={{ width: 160 }}
        />
      ),
    },
    {
      name: 'isTrashed',
      label: t('files.statusFilter'),
      children: (
        <Select
          allowClear
          placeholder={t('files.allStatus')}
          options={STATUS_OPTIONS}
          style={{ width: 160 }}
        />
      ),
    },
  ];

  return (
    <PageContainer title={t('files.title')}>
      <SearchForm
        fields={searchFields}
        onSearch={handleSearch}
        onReset={resetParams}
        initialValues={{
          search: params.search || undefined,
          fileCategory: params.fileCategory || undefined,
          isTrashed: params.isTrashed || undefined,
        }}
      />
      <ProTable<AdminFile>
        rowKey="id"
        columns={columns}
        queryResult={queryResult}
        pagination={{ current: params.page, pageSize: params.pageSize }}
        onChange={(pag) => {
          if (pag.current && pag.pageSize) {
            handlePageChange(pag.current, pag.pageSize);
          }
        }}
      />
    </PageContainer>
  );
}