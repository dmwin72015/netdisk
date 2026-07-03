import { Card, Row, Col, Spin, Result, Progress } from "antd";
import { useStorageStats } from "@/api/storage";
import PageContainer from "@/components/PageContainer";
import { formatBytes } from "@/utils/format";

const CATEGORY_COLORS: Record<string, string> = {
  video: "#722ed1",
  audio: "#13c2c2",
  image: "#1890ff",
  document: "#52c41a",
  archive: "#fa8c16",
  other: "#eb2f96",
  trash: "#ff4d4f",
};
const CATEGORY_LABELS: Record<string, string> = {
  video: "视频",
  audio: "音频",
  image: "图片",
  document: "文档",
  archive: "压缩包",
  other: "其他",
  trash: "回收站",
};

export default function StoragePage() {
  const { data: categories, isLoading, error } = useStorageStats();

  if (isLoading) {
    return (
      <div className="text-center p-15">
        <Spin size="large" />
      </div>
    );
  }
  if (error) {
    return (
      <PageContainer>
        <Result status="error" title="加载失败" subTitle={error.message} />
      </PageContainer>
    );
  }
  if (!categories || categories.length === 0) {
    return (
      <PageContainer>
        <Result status="warning" title="暂无数据" />
      </PageContainer>
    );
  }

  const totalBytes = categories.reduce((sum, c) => sum + c.bytes, 0);
  const trashCategory = categories.find((c) => c.category === "trash");
  const regularCategories = categories.filter((c) => c.category !== "trash");

  return (
    <PageContainer title="存储概览">
      <Row gutter={[16, 16]}>
        {regularCategories.map((stat) => {
          const pct =
            totalBytes > 0 ? Math.round((stat.bytes / totalBytes) * 100) : 0;
          const color = CATEGORY_COLORS[stat.category] || "#1890ff";
          return (
            <Col xs={24} sm={12} lg={8} key={stat.category}>
              <Card
                variant="borderless"
                title={CATEGORY_LABELS[stat.category] || stat.category}
              >
                <div className="mb-2">{stat.count} 个文件</div>
                <div className="mb-2 text-lg font-semibold" style={{ color }}>
                  {formatBytes(stat.bytes)}
                </div>
                <Progress
                  percent={pct}
                  strokeColor={color}
                  size="small"
                  showInfo={false}
                />
                <div className="text-right text-xs text-[#999] mt-1">
                  占比 {pct}%
                </div>
              </Card>
            </Col>
          );
        })}
      </Row>

      {trashCategory && (
        <Card title="回收站" className="mt-6!" variant="borderless">
          <Row gutter={16} align="middle">
            <Col>
              <span className="text-lg font-semibold text-[#ff4d4f]">
                {formatBytes(trashCategory.bytes)}
              </span>
            </Col>
            <Col>
              <span className="text-[#999]">{trashCategory.count} 个文件</span>
            </Col>
            <Col flex="auto">
              <Progress
                percent={
                  totalBytes > 0
                    ? Math.round((trashCategory.bytes / totalBytes) * 100)
                    : 0
                }
                strokeColor={CATEGORY_COLORS.trash}
                size="small"
              />
            </Col>
          </Row>
        </Card>
      )}
    </PageContainer>
  );
}
