import { useEffect, useState, useMemo } from 'react';
import { Table, Card, Switch, InputNumber, Button, message, Space, Typography, Input } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { adminListStorageStats, adminGetSettings, adminUpdateSettings, type AdminSettings } from '../api/admin';

const { Title } = Typography;

function formatBytes(b: number): string {
  if (b === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(b) / Math.log(k));
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function parseBytesInput(val: number): number {
  return Math.round(val * 1024 * 1024 * 1024);
}

export default function Storage() {
  const [items, setItems] = useState<{ id: string; username: string; usedBytes: number; totalBytes: number }[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState<'used' | 'name'>('used');
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<AdminSettings | null>(null);
  const [formValues, setFormValues] = useState({ siteName: '', allowRegistration: true, defaultQuota: 0, maxUploadSize: 0 });

  useEffect(() => {
    setLoading(true);
    adminListStorageStats(100)
      .then((res) => {
        setItems(res.items);
        setTotal(res.total);
      })
      .catch(() => message.error('Failed to load storage stats'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    adminGetSettings()
      .then((s) => {
        setSettings(s);
        setFormValues({
          siteName: s.siteName,
          allowRegistration: s.allowRegistration,
          defaultQuota: s.defaultQuota,
          maxUploadSize: s.maxUploadSize,
        });
      })
      .catch(() => message.error('Failed to load settings'));
  }, []);

  const sortedItems = useMemo(() => {
    const arr = [...items];
    if (sortBy === 'used') {
      arr.sort((a, b) => b.usedBytes - a.usedBytes);
    } else {
      arr.sort((a, b) => a.username.localeCompare(b.username));
    }
    return arr;
  }, [items, sortBy]);

  const maxUsed = Math.max(...items.map((s) => s.usedBytes), 1);

  const handleSaveSettings = async () => {
    setSaving(true);
    try {
      const updated = await adminUpdateSettings({
        siteName: formValues.siteName,
        allowRegistration: formValues.allowRegistration,
        defaultQuota: parseBytesInput(formValues.defaultQuota),
        maxUploadSize: parseBytesInput(formValues.maxUploadSize),
      });
      setSettings(updated);
      message.success('Settings saved');
    } catch {
      message.error('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const columns = [
    { title: 'Username', dataIndex: 'username', key: 'username', sorter: (a: typeof items[0], b: typeof items[0]) => a.username.localeCompare(b.username) },
    {
      title: 'Used',
      dataIndex: 'usedBytes',
      key: 'usedBytes',
      sorter: (a: typeof items[0], b: typeof items[0]) => b.usedBytes - a.usedBytes,
      render: (v: number) => formatBytes(v),
    },
    {
      title: 'Total',
      dataIndex: 'totalBytes',
      key: 'totalBytes',
      render: (v: number) => formatBytes(v),
    },
    {
      title: 'Usage',
      key: 'usage',
      render: (_: unknown, record: { usedBytes: number; totalBytes: number }) => {
        const pct = record.totalBytes > 0 ? Math.round((record.usedBytes / record.totalBytes) * 100) : 0;
        return (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <div style={{ flex: 1, height: 8, background: '#f5f5f5', borderRadius: 4, overflow: 'hidden' }}>
              <div
                style={{
                  height: '100%',
                  width: `${pct}%`,
                  background: pct > 80 ? '#ff4d4f' : '#1890ff',
                  borderRadius: 4,
                }}
              />
            </div>
            <span style={{ minWidth: 40, textAlign: 'right' }}>{pct}%</span>
          </div>
        );
      },
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>
          Storage Statistics
          {sortBy === 'used' ? ' (sorted by usage)' : ' (sorted by name)'}
        </Title>
        <Space>
          <span style={{ color: '#666' }}>Sort:</span>
          <Button
            type={sortBy === 'used' ? 'primary' : 'default'}
            size="small"
            onClick={() => setSortBy('used')}
          >
            By Usage
          </Button>
          <Button
            type={sortBy === 'name' ? 'primary' : 'default'}
            size="small"
            onClick={() => setSortBy('name')}
          >
            By Name
          </Button>
        </Space>
      </div>

      <Card title="Top 20 Storage Usage" style={{ marginBottom: 24 }}>
        {sortedItems.slice(0, 20).length === 0 && (
          <div style={{ textAlign: 'center', padding: 20, color: '#999' }}>No data</div>
        )}
        {sortedItems.slice(0, 20).map((s) => {
          const pct = (s.usedBytes / maxUsed) * 100;
          const usagePct = s.totalBytes > 0 ? Math.round((s.usedBytes / s.totalBytes) * 100) : 0;
          return (
            <div key={s.id} style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                <span>{s.username}</span>
                <span style={{ color: '#666', fontSize: 12 }}>
                  {formatBytes(s.usedBytes)} / {formatBytes(s.totalBytes)} ({usagePct}%)
                </span>
              </div>
              <div style={{ height: 18, background: '#f5f5f5', borderRadius: 4, overflow: 'hidden' }}>
                <div
                  style={{
                    height: '100%',
                    width: `${pct}%`,
                    background: usagePct > 80 ? '#ff4d4f' : '#1890ff',
                    borderRadius: 4,
                    transition: 'width 0.3s',
                    minWidth: usagePct > 0 ? 2 : 0,
                  }}
                />
              </div>
            </div>
          );
        })}
      </Card>

      <Card title={`All Users (${total})`}>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={sortedItems}
          loading={loading}
          pagination={{ pageSize: 20, showTotal: (t) => `Total ${t}` }}
          size="small"
        />
      </Card>

      <Card title="System Settings" style={{ marginTop: 24 }}>
        {settings && (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Space style={{ width: '100%', justifyContent: 'space-between' }}>
              <span>Allow Registration</span>
              <Switch
                checked={formValues.allowRegistration}
                onChange={(v) => setFormValues((f) => ({ ...f, allowRegistration: v }))}
              />
            </Space>
            <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap', width: '100%' }}>
              <div style={{ flex: 1, minWidth: 200 }}>
                <label style={{ display: 'block', marginBottom: 4, color: '#666' }}>Site Name</label>
                <Input
                  value={formValues.siteName}
                  onChange={(e) => setFormValues((f) => ({ ...f, siteName: e.target.value }))}
                />
              </div>
              <div style={{ flex: 1, minWidth: 200 }}>
                <label style={{ display: 'block', marginBottom: 4, color: '#666' }}>Default Quota (GB)</label>
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  value={formValues.defaultQuota}
                  onChange={(v) => setFormValues((f) => ({ ...f, defaultQuota: v ?? 0 }))}
                />
              </div>
              <div style={{ flex: 1, minWidth: 200 }}>
                <label style={{ display: 'block', marginBottom: 4, color: '#666' }}>Max Upload Size (GB)</label>
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  value={formValues.maxUploadSize}
                  onChange={(v) => setFormValues((f) => ({ ...f, maxUploadSize: v ?? 0 }))}
                />
              </div>
            </div>
            <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={handleSaveSettings}>
              Save Settings
            </Button>
          </Space>
        )}
      </Card>
    </div>
  );
}
