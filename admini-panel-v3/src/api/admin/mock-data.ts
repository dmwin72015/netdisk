import type {
  AdminUser,
  AdminFile,
  AdminPhysicalFile,
  AdminDashboardStats,
  CategoryStat,
  SystemConfigItem,
  AdminActivityLog,
  AdminActionLabel,
} from './types';

// ---- Helpers ----
const now = Math.floor(Date.now() / 1000);
const randomOffset = (maxDays: number) => Math.floor(Math.random() * maxDays * 86400);
const randomId = () => Math.random().toString(36).slice(2, 10);
const randomSlug = () => Math.random().toString(36).slice(2, 12) + Math.random().toString(36).slice(2, 12);
const randomHash = () =>
  Array.from({ length: 64 }, () => Math.floor(Math.random() * 16).toString(16)).join('');

// ---- Mock Users (20) ----
const userNames = [
  '张伟', '王芳', '李强', '刘洋', '陈静',
  '杨磊', '赵敏', '黄勇', '周婷', '吴刚',
  '徐慧', '孙鹏', '马丽', '朱军', '胡明',
  '郭靖', '林燕', '何平', '高洁', '罗辉',
];

const emails = userNames.map((_, i) => `user${i + 1}@example.com`);
const roles = ['admin', 'user', 'user', 'user', 'vip', 'user', 'user', 'vip', 'user', 'user',
  'user', 'user', 'vip', 'user', 'user', 'user', 'user', 'user', 'user', 'user'];
const registerMethods = ['email', 'google', 'github', 'email', 'google', 'email', 'email', 'github', 'email', 'email',
  'google', 'email', 'email', 'email', 'github', 'email', 'email', 'email', 'google', 'email'];
const statuses = [1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1];

export const mockUsers: AdminUser[] = userNames.map((name, i) => {
  const baseBytes = [5368709120, 10737418240, 5368709120, 21474836480, 10737418240][i % 5];
  const memberBonusBytes = i % 3 === 0 ? 5368709120 : 0;
  const packBytes = i % 4 === 0 ? 2147483648 : 0;
  const totalBytes = baseBytes + memberBonusBytes + packBytes;
  const usedBytes = Math.floor(totalBytes * (0.1 + Math.random() * 0.7));

  return {
    id: String(i + 1),
    slug: randomSlug(),
    username: name,
    email: emails[i],
    role: roles[i],
    registerMethod: registerMethods[i],
    status: statuses[i],
    usedBytes,
    baseBytes,
    memberBonusBytes,
    packBytes,
    totalBytes,
    createdAt: now - randomOffset(365),
    profile: {
      displayName: name,
      avatarUrl: `https://api.dicebear.com/7.x/avataaars/svg?seed=${i}`,
      bio: i % 3 === 0 ? '这是我的个人简介' : '',
    },
    oauthAccounts: registerMethods[i] !== 'email'
      ? [{ provider: registerMethods[i], providerAccountId: `oauth_${i}_${randomId()}`, createdAt: now - randomOffset(365) }]
      : [],
  };
});

// ---- Mock Files (50) ----
const fileCategories = ['video', 'audio', 'image', 'document', 'archive', 'other'];
const mimeTypes: Record<string, string[]> = {
  video: ['video/mp4', 'video/mkv', 'video/avi', 'video/webm'],
  audio: ['audio/mp3', 'audio/flac', 'audio/wav', 'audio/aac'],
  image: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  document: ['application/pdf', 'application/msword', 'text/plain', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'],
  archive: ['application/zip', 'application/x-rar-compressed', 'application/x-7z-compressed'],
  other: ['application/octet-stream'],
};

const fileNames: Record<string, string[]> = {
  video: ['电影_流浪地球.mp4', '纪录片_蓝色星球.mkv', '教程_React入门.mp4', '短视频_旅行vlog.mp4', '会议录像_Q3.webm'],
  audio: ['周杰伦_晴天.mp3', '播客_技术漫谈.flac', '录音_会议记录.wav', '音乐_夜曲.mp3', '有声书_三体.aac'],
  image: ['照片_风景.jpg', '截图_2024-01.png', '头像.gif', '设计稿_首页.webp', '扫描件_合同.jpg'],
  document: ['项目计划书.pdf', '需求文档_v2.docx', '笔记.txt', '财务报表.xlsx', '技术架构设计.pdf'],
  archive: ['备份_2024.zip', '项目源码.rar', '素材包.7z', '资料汇总.zip'],
  other: ['数据文件.bin', '未知文件.dat'],
};

export const mockFiles: AdminFile[] = Array.from({ length: 50 }, (_, i) => {
  const category = fileCategories[i % fileCategories.length];
  const names = fileNames[category];
  const fileName = names[i % names.length];
  const mimes = mimeTypes[category];
  const mimeType = mimes[i % mimes.length];
  const userId = String((i % 20) + 1);
  const user = mockUsers[Number(userId) - 1];

  let fileSize: number;
  switch (category) {
    case 'video': fileSize = Math.floor(100 * 1024 * 1024 + Math.random() * 4 * 1024 * 1024 * 1024); break;
    case 'audio': fileSize = Math.floor(3 * 1024 * 1024 + Math.random() * 50 * 1024 * 1024); break;
    case 'image': fileSize = Math.floor(100 * 1024 + Math.random() * 20 * 1024 * 1024); break;
    case 'document': fileSize = Math.floor(10 * 1024 + Math.random() * 50 * 1024 * 1024); break;
    case 'archive': fileSize = Math.floor(1024 * 1024 + Math.random() * 500 * 1024 * 1024); break;
    default: fileSize = Math.floor(1024 + Math.random() * 1024 * 1024); break;
  }

  const created = now - randomOffset(180);
  return {
    id: String(1000 + i),
    userId,
    slug: randomSlug(),
    username: user.username,
    fileName,
    isDir: false,
    fileSize,
    mimeType,
    fileCategory: category,
    isTrashed: i % 12 === 0,
    isStarred: i % 7 === 0,
    createdAt: created,
    updatedAt: created + Math.floor(Math.random() * 86400 * 7),
  };
});

// ---- Mock Physical Files (30) ----
const physicalStatuses = ['active', 'active', 'active', 'orphaned', 'pending_delete'];

export const mockPhysicalFiles: AdminPhysicalFile[] = Array.from({ length: 30 }, (_, i) => {
  const status = physicalStatuses[i % physicalStatuses.length];
  const category = fileCategories[i % fileCategories.length];
  const mimes = mimeTypes[category];
  const mimeType = mimes[i % mimes.length];
  const userFileCount = status === 'orphaned' ? 0 : Math.floor(1 + Math.random() * 5);
  const mediaItemCount = status === 'orphaned' ? 0 : Math.floor(Math.random() * 3);

  let fileSize: number;
  switch (category) {
    case 'video': fileSize = Math.floor(500 * 1024 * 1024 + Math.random() * 5 * 1024 * 1024 * 1024); break;
    case 'audio': fileSize = Math.floor(5 * 1024 * 1024 + Math.random() * 80 * 1024 * 1024); break;
    case 'image': fileSize = Math.floor(500 * 1024 + Math.random() * 30 * 1024 * 1024); break;
    case 'document': fileSize = Math.floor(50 * 1024 + Math.random() * 100 * 1024 * 1024); break;
    case 'archive': fileSize = Math.floor(10 * 1024 * 1024 + Math.random() * 1024 * 1024 * 1024); break;
    default: fileSize = Math.floor(1024 + Math.random() * 500 * 1024 * 1024); break;
  }

  const hash = randomHash();
  return {
    id: 2000 + i,
    slug: randomSlug(),
    hashAlgo: 'sha256',
    fileHash: hash,
    preHash: hash.slice(0, 16),
    fileSize,
    mimeType,
    storagePath: `/data/storage/${hash.slice(0, 2)}/${hash.slice(2, 4)}/${hash}`,
    status,
    createdAt: now - randomOffset(300),
    userFileCount,
    mediaItemCount,
  };
});

// ---- Mock Dashboard Stats ----
export const mockDashboardStats: AdminDashboardStats = {
  totalUsers: 1258,
  totalFiles: 89432,
  totalStorage: 109951162777600, // 100TB
  storageUsed: 45097156608000,  // ~41TB
  newTodayUsers: 23,
  newTodayFiles: 456,
  diskTotal: 109951162777600,
  diskUsed: 45097156608000,
  diskFree: 64854006169600,
};

// ---- Mock Storage Category Stats ----
export const mockCategoryStats: CategoryStat[] = [
  { category: 'video', bytes: 28991029248000, count: 34521 },
  { category: 'audio', bytes: 3221225472000, count: 18200 },
  { category: 'image', bytes: 5368709120000, count: 22150 },
  { category: 'document', bytes: 2147483648000, count: 8432 },
  { category: 'archive', bytes: 4294967296000, count: 4200 },
  { category: 'other', bytes: 858993459200, count: 1129 },
  { category: 'trash', bytes: 214748364800, count: 800 },
];

// ---- Mock System Config ----
export const mockSystemConfig: SystemConfigItem[] = [
  { key: 'max_upload_size', value: '5368709120', defaultValue: '5368709120', description: '单文件最大上传大小（字节），默认 5GB' },
  { key: 'default_storage_quota', value: '5368709120', defaultValue: '5368709120', description: '新用户默认存储配额（字节），默认 5GB' },
  { key: 'member_bonus_storage', value: '5368709120', defaultValue: '5368709120', description: '会员额外存储（字节），默认 5GB' },
  { key: 'allow_registration', value: 'true', defaultValue: 'true', description: '是否允许新用户注册' },
  { key: 'allow_google_login', value: 'true', defaultValue: 'true', description: '是否允许 Google 登录' },
  { key: 'allow_github_login', value: 'true', defaultValue: 'true', description: '是否允许 GitHub 登录' },
  { key: 'recycle_bin_days', value: '30', defaultValue: '30', description: '回收站保留天数' },
  { key: 'upload_chunk_size', value: '5242880', defaultValue: '5242880', description: '分片上传每片大小（字节），默认 5MB' },
  { key: 'site_name', value: 'NetDisk', defaultValue: 'NetDisk', description: '站点名称' },
  { key: 'maintenance_mode', value: 'false', defaultValue: 'false', description: '是否启用维护模式' },
];

// ---- Mock Activity Logs (20) ----
const actions = [
  { action: 'user.login', actionLabel: '用户登录' },
  { action: 'user.logout', actionLabel: '用户登出' },
  { action: 'file.upload', actionLabel: '上传文件' },
  { action: 'file.download', actionLabel: '下载文件' },
  { action: 'file.delete', actionLabel: '删除文件' },
  { action: 'file.restore', actionLabel: '恢复文件' },
  { action: 'file.star', actionLabel: '收藏文件' },
  { action: 'file.move', actionLabel: '移动文件' },
  { action: 'folder.create', actionLabel: '创建文件夹' },
  { action: 'user.update_profile', actionLabel: '更新个人资料' },
];

const resourceTypes = ['file', 'folder', 'user', 'system'];
const ips = ['192.168.1.100', '10.0.0.52', '172.16.0.88', '223.104.3.200', '101.226.103.56', '36.152.44.109', '119.6.136.79', '183.192.200.10'];
const ipRegions = ['上海', '北京', '广州', '深圳', '杭州', '成都', '武汉', '南京'];
const browsers = ['Chrome 120', 'Firefox 121', 'Safari 17', 'Edge 120', 'Chrome 119'];
const oses = ['Windows 11', 'macOS 14', 'Ubuntu 22.04', 'iOS 17', 'Android 14'];
const userAgents = [
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
  'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15',
  'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15',
  'Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36',
];

export const mockActivityLogs: AdminActivityLog[] = Array.from({ length: 20 }, (_, i) => {
  const actionEntry = actions[i % actions.length];
  const userId = (i % 20) + 1;
  const user = mockUsers[userId - 1];
  const regionIdx = i % ipRegions.length;

  return {
    id: 3000 + i,
    userId,
    username: user.username,
    action: actionEntry.action,
    actionLabel: actionEntry.actionLabel,
    resourceType: resourceTypes[i % resourceTypes.length],
    resourceName: i % 2 === 0 ? `文件_${i}.mp4` : `文件夹_${i}`,
    ip: ips[regionIdx],
    ipRegion: ipRegions[regionIdx],
    userAgent: userAgents[i % userAgents.length],
    os: oses[i % oses.length],
    browser: browsers[i % browsers.length],
    extra: i % 3 === 0 ? { detail: '操作详情信息' } : null,
    createdAt: new Date((now - randomOffset(30)) * 1000).toISOString(),
  };
});

// ---- Mock Action Labels ----
export const mockActionLabels: AdminActionLabel[] = actions.map(a => ({
  action: a.action,
  label: a.actionLabel,
}));
