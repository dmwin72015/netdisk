import {
  Spin,
  Card,
  Row,
  Col,
  Tag,
  Avatar,
  Button,
  Result,
  Statistic,
  Divider,
} from "antd";
import { ProDescriptions } from "@ant-design/pro-components";
import { ArrowLeftOutlined, UserOutlined } from "@ant-design/icons";
import { useNavigate, useParams } from "react-router";
import { useUser } from "@/api/users";
import PageContainer from "@/components/PageContainer";
import { formatBytes, formatDate } from "@/utils/format";
import CopyCell from "@/components/CopyCell";

const ROLE_COLORS: Record<string, string> = {
  admin: "red",
  user: "blue",
  vip: "orange",
};

export default function UserDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: user, isLoading, error } = useUser(id!);

  if (isLoading) {
    return (
      <PageContainer title="用户详情">
        <div className="text-center p-15">
          <Spin size="large" />
        </div>
      </PageContainer>
    );
  }

  if (error) {
    return (
      <PageContainer title="用户详情">
        <Result
          status="error"
          title="加载失败"
          subTitle={error.message}
          extra={
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate("/users")}
            >
              返回
            </Button>
          }
        />
      </PageContainer>
    );
  }

  if (!user) {
    return (
      <PageContainer title="用户详情">
        <Result
          status="warning"
          title="用户不存在"
          extra={
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate("/users")}
            >
              返回
            </Button>
          }
        />
      </PageContainer>
    );
  }

  const usagePct =
    user.totalBytes > 0
      ? Math.round((user.usedBytes / user.totalBytes) * 100)
      : 0;

  return (
    <PageContainer
      title={`用户详情: ${user.username}`}
      extra={[
        <Button
          key="back"
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate("/users")}
        >
          返回列表
        </Button>,
      ]}
    >
      <Row gutter={16}>
        <Col span={6}>
          <Card hoverable>
            <Statistic title="已用存储" value={formatBytes(user.usedBytes)} />
          </Card>
        </Col>
        <Col span={6}>
          <Card hoverable>
            <Statistic title="总配额" value={formatBytes(user.totalBytes)} />
          </Card>
        </Col>
        <Col span={6}>
          <Card hoverable>
            <Statistic
              title="使用率"
              value={usagePct}
              suffix="%"
              valueStyle={{
                color:
                  usagePct > 80
                    ? "#ff4d4f"
                    : usagePct > 60
                      ? "#faad14"
                      : "#1890ff",
              }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card hoverable>
            <Statistic title="注册时间" value={formatDate(user.createdAt)} />
          </Card>
        </Col>
      </Row>

      <Card className="mt-6!" variant="borderless">
        <ProDescriptions title="基本信息" column={2} bordered size="small">
          <ProDescriptions.Item label="用户ID">{user.id}</ProDescriptions.Item>
          <ProDescriptions.Item label="用户名">
            {user.username}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="Slug">
            <CopyCell value={user.slug} />
          </ProDescriptions.Item>
          <ProDescriptions.Item label="邮箱">{user.email}</ProDescriptions.Item>
          <ProDescriptions.Item label="角色">
            <Tag color={ROLE_COLORS[user.role] || "default"}>{user.role}</Tag>
          </ProDescriptions.Item>
          <ProDescriptions.Item label="注册方法">
            {user.registerMethod}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="状态">
            {user.status === 1 ? "正常" : "禁用"}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="注册时间">
            {formatDate(user.createdAt)}
          </ProDescriptions.Item>
        </ProDescriptions>

        {user.profile && (
          <>
            <Divider />
            <ProDescriptions title="个人资料" column={2} bordered size="small">
              {user.profile.displayName && (
                <ProDescriptions.Item label="昵称">
                  {user.profile.displayName}
                </ProDescriptions.Item>
              )}
              {user.profile.avatarUrl && (
                <ProDescriptions.Item label="头像">
                  <Avatar
                    src={user.profile.avatarUrl}
                    icon={<UserOutlined />}
                  />
                </ProDescriptions.Item>
              )}
              {user.profile.bio && (
                <ProDescriptions.Item label="简介" span={2}>
                  {user.profile.bio}
                </ProDescriptions.Item>
              )}
            </ProDescriptions>
          </>
        )}

        <Divider />
        <ProDescriptions title="存储信息" column={2} bordered size="small">
          <ProDescriptions.Item label="已用存储">
            {formatBytes(user.usedBytes)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="总配额">
            {formatBytes(user.totalBytes)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="基础配额">
            {formatBytes(user.baseBytes)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="会员赠送">
            {formatBytes(user.memberBonusBytes)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="购买容量">
            {formatBytes(user.packBytes)}
          </ProDescriptions.Item>
          <ProDescriptions.Item label="使用率">
            <div className="flex items-center gap-2">
              <div className="flex-1 h-2 bg-[#f5f5f5] rounded overflow-hidden">
                <div
                  className="h-full rounded transition-all"
                  style={{
                    width: `${usagePct}%`,
                    background:
                      usagePct > 80
                        ? "#ff4d4f"
                        : usagePct > 60
                          ? "#faad14"
                          : "#1890ff",
                  }}
                />
              </div>
              <span className="text-sm">{usagePct}%</span>
            </div>
          </ProDescriptions.Item>
        </ProDescriptions>
      </Card>

      {user.oauthAccounts && user.oauthAccounts.length > 0 && (
        <Card title="OAuth 账户" className="mt-6!" variant="borderless">
          {user.oauthAccounts.map((acc) => (
            <div
              key={acc.provider}
              className="flex items-center gap-3 py-2 border-b border-[#f0f0f0] last:border-b-0"
            >
              <Tag color="blue" className="!mr-0">
                {acc.provider}
              </Tag>
              <span className="font-mono text-xs text-[#666]">
                {acc.providerAccountId}
              </span>
              <span className="text-xs text-[#999] ml-auto">
                {formatDate(acc.createdAt)}
              </span>
            </div>
          ))}
        </Card>
      )}
    </PageContainer>
  );
}
