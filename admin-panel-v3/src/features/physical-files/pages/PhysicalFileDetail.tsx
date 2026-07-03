import { useParams, Link } from "react-router";
import {
  Button,
  Card,
  Col,
  Result,
  Row,
  Spin,
  Statistic,
  Table,
  Tag,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { ArrowLeftOutlined } from "@ant-design/icons";
import { ProDescriptions } from "@ant-design/pro-components";
import { PageContainer } from "@/components/PageContainer";
import CopyCell from "@/components/CopyCell";
import { usePhysicalFileDetail } from "@/api/physical-files";
import type { CleanupQueryUserFile } from "@/api/cleanup";
import { formatBytes, formatDate } from "@/utils/format";

const STATUS_COLORS: Record<string, string> = {
  completed: "green",
  processing: "blue",
  failed: "red",
};

export default function PhysicalFileDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading, isError, error } = usePhysicalFileDetail(id!);

  const userFileColumns: ColumnsType<CleanupQueryUserFile> = [
    { title: "ID", dataIndex: "id", width: 80 },
    { title: "Slug", dataIndex: "slug", width: 180, ellipsis: true },
    { title: "文件名", dataIndex: "fileName", width: 200, ellipsis: true },
    {
      title: "用户",
      key: "user",
      width: 120,
      render: (_, record) => (
        <Link to={`/users/${record.userId}`}>{record.username}</Link>
      ),
    },
    {
      title: "大小",
      dataIndex: "fileSize",
      width: 120,
      render: (_, record) => formatBytes(record.fileSize),
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 170,
      render: (_, record) => formatDate(record.createdAt),
    },
  ];

  if (isLoading) {
    return (
      <PageContainer title="物理文件详情">
        <div className="text-center p-15">
          <Spin size="large" />
        </div>
      </PageContainer>
    );
  }

  if (isError || !data) {
    return (
      <PageContainer title="物理文件详情">
        <Result
          status="error"
          title="加载失败"
          subTitle={error instanceof Error ? error.message : ""}
          extra={
            <Link to="/files/physical">
              <Button icon={<ArrowLeftOutlined />}>返回</Button>
            </Link>
          }
        />
      </PageContainer>
    );
  }

  const { physicalFile, userFiles, totalUploads, uniqueUsers } = data;
  const totalSize = userFiles.reduce((a, b) => a + b.fileSize, 0);

  return (
    <PageContainer
      title="物理文件详情"
      extra={
        <Link to="/files/physical">
          <Button icon={<ArrowLeftOutlined />}>返回列表</Button>
        </Link>
      }
    >
      <Row gutter={16}>
        <Col span={6}>
          <Card variant="borderless">
            <Statistic title="总上传数" value={totalUploads} />
          </Card>
        </Col>
        <Col span={6}>
          <Card variant="borderless">
            <Statistic title="独立用户数" value={uniqueUsers} />
          </Card>
        </Col>
        <Col span={6}>
          <Card variant="borderless">
            <Statistic title="总大小" value={formatBytes(totalSize)} />
          </Card>
        </Col>
        <Col span={6}>
          <Card variant="borderless">
            <Statistic
              title="引用次数"
              value={physicalFile.userFileCount + physicalFile.mediaItemCount}
            />
          </Card>
        </Col>
      </Row>

      <Card title="物理文件信息" className="mt-6!" variant="borderless">
        <ProDescriptions column={2} bordered size="small">
          <ProDescriptions.Item label="ID">
            {physicalFile.id}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="Slug">
            <CopyCell value={physicalFile.slug} />
          </ProDescriptions.Item>
          <ProDescriptions.Item label="文件哈希" span={2}>
            <CopyCell value={physicalFile.fileHash}>
              <code className="text-xs">{physicalFile.fileHash}</code>
            </CopyCell>
          </ProDescriptions.Item>
          <ProDescriptions.Item label="文件大小">
            {formatBytes(physicalFile.fileSize)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="MIME 类型">
            {physicalFile.mimeType}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="存储路径" span={2}>
            <code className="text-xs">{physicalFile.storagePath}</code>
          </ProDescriptions.Item>
          <ProDescriptions.Item label="完整路径" span={2}>
            <code className="text-xs">{data.fullPath}</code>
          </ProDescriptions.Item>
          <ProDescriptions.Item label="哈希算法">
            {physicalFile.hashAlgo}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="状态">
            <Tag color={STATUS_COLORS[physicalFile.status] || "default"}>
              {physicalFile.status}
            </Tag>
          </ProDescriptions.Item>
          <ProDescriptions.Item label="创建时间">
            {formatDate(physicalFile.createdAt)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="引用次数">
            {physicalFile.userFileCount + physicalFile.mediaItemCount}
          </ProDescriptions.Item>
        </ProDescriptions>
      </Card>

      {userFiles.length > 0 && (
        <Card title="用户上传记录" className="mt-6!" variant="borderless">
          <Table<CleanupQueryUserFile>
            rowKey="id"
            columns={userFileColumns}
            dataSource={userFiles}
            pagination={false}
            size="small"
            scroll={{ x: "max-content" }}
          />
        </Card>
      )}
    </PageContainer>
  );
}
