import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Button, Card, Form, Input, Layout, Menu, Select, Space, Table, Tabs, Tag, Tree, Typography, message } from 'antd';
import { ApiOutlined, DashboardOutlined, LockOutlined, SettingOutlined, TeamOutlined } from '@ant-design/icons';
import { api } from './api/client';
import './styles/app.css';

const { Header, Sider, Content } = Layout;

type ConfigValue = {
  key: string;
  value: string;
  source: string;
  version: number;
  updated_at: string;
};

type ConfigDefinition = {
  key: string;
  type: string;
  default_value: string;
  editable: boolean;
  sensitive: boolean;
  description: string;
};

type ConfigPayload = {
  definitions: ConfigDefinition[];
  values: ConfigValue[];
};

type ConfigLog = {
  id: string;
  key: string;
  old_value: string;
  new_value: string;
  action: string;
  version: number;
  changed_by: string;
  created_at: string;
};

type TokenResponse = {
  access_token: string;
  refresh_token: string;
  access_token_expires_at: string;
  refresh_token_expires_at: string;
  user: { email: string; display_name: string; tenant_id: string };
  tenant: { id: string; name: string };
};

type OrganizationNode = {
  id: string;
  name: string;
  org_type: string;
  status: string;
  parent_id?: string;
  children: OrganizationNode[];
};

type Team = {
  id: string;
  organization_id: string;
  name: string;
  status: string;
};

type Channel = {
  id: string;
  organization_id: string;
  platform: string;
  display_name: string;
  status: string;
};

function App() {
  const [active, setActive] = useState('config');

  return (
    <Layout className="app-shell">
      <Sider width={232} theme="light" className="app-sider">
        <div className="brand">EduAdCRM</div>
        <Menu
          mode="inline"
          selectedKeys={[active]}
          onClick={(item) => setActive(item.key)}
          items={[
            { key: 'config', icon: <SettingOutlined />, label: '系统配置' },
            { key: 'auth', icon: <LockOutlined />, label: '用户鉴权' },
            { key: 'control', icon: <TeamOutlined />, label: '组织配置' },
            { key: 'oauth', icon: <ApiOutlined />, label: '广告授权' },
            { key: 'overview', icon: <DashboardOutlined />, label: '平台概览' }
          ]}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <Typography.Title level={4}>教育行业广告 CRM 数据分析平台</Typography.Title>
        </Header>
        <Content className="app-content">
          {active === 'config' && <ConfigPage />}
          {active === 'auth' && <AuthPage />}
          {active === 'control' && <ControlPlanePage />}
          {active === 'oauth' && <OAuthPage />}
          {active === 'overview' && <OverviewPage />}
        </Content>
      </Layout>
    </Layout>
  );
}

function ControlPlanePage() {
  const [orgs, setOrgs] = useState<OrganizationNode[]>([]);
  const [teams, setTeams] = useState<Team[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);

  const load = async () => {
    setOrgs(await api<OrganizationNode[]>('/api/orgs/tree'));
    setTeams(await api<Team[]>('/api/teams'));
    setChannels(await api<Channel[]>('/api/channels'));
  };

  useEffect(() => {
    void load().catch(() => undefined);
  }, []);

  const treeData = orgs.map(toTreeNode);

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <Typography.Title level={3}>组织配置</Typography.Title>
      <Card title="创建组织">
        <Form layout="inline" onFinish={async (values) => {
          await api('/api/orgs', { method: 'POST', body: JSON.stringify(values) });
          message.success('组织已创建');
          await load();
        }}>
          <Form.Item name="name" rules={[{ required: true }]}><Input placeholder="组织名称" /></Form.Item>
          <Form.Item name="org_type" initialValue="school">
            <Select
              style={{ width: 140 }}
              options={[
                { value: 'group', label: '集团' },
                { value: 'region', label: '区域' },
                { value: 'school', label: '学校' },
                { value: 'campus', label: '校区' }
              ]}
            />
          </Form.Item>
          <Form.Item name="parent_id"><Input placeholder="父组织 ID，可空" /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="组织树">
        <Tree treeData={treeData} defaultExpandAll />
      </Card>
      <Card title="创建团队">
        <Form layout="inline" onFinish={async (values) => {
          await api('/api/teams', { method: 'POST', body: JSON.stringify(values) });
          message.success('团队已创建');
          await load();
        }}>
          <Form.Item name="organization_id" rules={[{ required: true }]}><Input placeholder="组织 ID" /></Form.Item>
          <Form.Item name="name" rules={[{ required: true }]}><Input placeholder="团队名称" /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="团队列表">
        <Table rowKey="id" dataSource={teams} pagination={false} columns={[
          { title: '团队 ID', dataIndex: 'id' },
          { title: '组织 ID', dataIndex: 'organization_id' },
          { title: '名称', dataIndex: 'name' },
          { title: '状态', dataIndex: 'status', render: (status) => <Tag>{status}</Tag> }
        ]} />
      </Card>
      <Card title="创建广告渠道">
        <Form layout="inline" onFinish={async (values) => {
          await api('/api/channels', { method: 'POST', body: JSON.stringify(values) });
          message.success('渠道已创建');
          await load();
        }}>
          <Form.Item name="organization_id" rules={[{ required: true }]}><Input placeholder="组织 ID" /></Form.Item>
          <Form.Item name="platform" initialValue="douyin">
            <Select
              style={{ width: 160 }}
              options={[
                { value: 'douyin', label: '抖音' },
                { value: 'tencent', label: '腾讯' },
                { value: 'baidu', label: '百度' },
                { value: 'xiaohongshu', label: '小红书' }
              ]}
            />
          </Form.Item>
          <Form.Item name="display_name"><Input placeholder="展示名称" /></Form.Item>
          <Button type="primary" htmlType="submit">创建</Button>
        </Form>
      </Card>
      <Card title="渠道列表">
        <Table rowKey="id" dataSource={channels} pagination={false} columns={[
          { title: '渠道 ID', dataIndex: 'id' },
          { title: '组织 ID', dataIndex: 'organization_id' },
          { title: '平台', dataIndex: 'platform' },
          { title: '展示名称', dataIndex: 'display_name' },
          { title: '状态', dataIndex: 'status', render: (status) => <Tag>{status}</Tag> }
        ]} />
      </Card>
    </Space>
  );
}

function toTreeNode(org: OrganizationNode): { key: string; title: string; children: ReturnType<typeof toTreeNode>[] } {
  return {
    key: org.id,
    title: `${org.name} (${org.org_type}) - ${org.id}`,
    children: (org.children || []).map(toTreeNode)
  };
}

function ConfigPage() {
  const [config, setConfig] = useState<ConfigPayload>({ definitions: [], values: [] });
  const [logs, setLogs] = useState<ConfigLog[]>([]);
  const [form] = Form.useForm();

  const load = async () => {
    setConfig(await api<ConfigPayload>('/api/config'));
    setLogs(await api<ConfigLog[]>('/api/config/change-logs'));
  };

  useEffect(() => {
    void load();
  }, []);

  const rows = config.values.map((value) => ({
    ...value,
    definition: config.definitions.find((item) => item.key === value.key)
  }));

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <Typography.Title level={3}>系统配置</Typography.Title>
      <Card title="发布配置">
        <Form
          form={form}
          layout="inline"
          onFinish={async (values) => {
            await api(`/api/config/${values.key}`, { method: 'PATCH', body: JSON.stringify({ value: values.value }) });
            message.success('配置已发布，后续新签发 token 将使用新值');
            form.resetFields();
            await load();
          }}
        >
          <Form.Item name="key" rules={[{ required: true }]} className="wide-select">
            <Select
              placeholder="选择配置项"
              options={config.definitions.filter((item) => item.editable).map((item) => ({ label: item.key, value: item.key }))}
            />
          </Form.Item>
          <Form.Item name="value" rules={[{ required: true }]}>
            <Input placeholder="例如 24h / 720h / true" />
          </Form.Item>
          <Button type="primary" htmlType="submit">发布</Button>
        </Form>
      </Card>
      <Card title="当前配置">
        <Table
          rowKey="key"
          dataSource={rows}
          pagination={false}
          columns={[
            { title: '配置键', dataIndex: 'key' },
            { title: '当前值', dataIndex: 'value' },
            { title: '来源', dataIndex: 'source', render: (source) => <Tag>{source}</Tag> },
            { title: '版本', dataIndex: 'version' },
            { title: '说明', render: (_, row) => row.definition?.description },
            {
              title: '操作',
              render: (_, row) => row.definition?.editable ? (
                <Button size="small" onClick={async () => {
                  await api(`/api/config/${row.key}/rollback`, { method: 'POST', body: JSON.stringify({}) });
                  message.success('已回滚配置');
                  await load();
                }}>回滚</Button>
              ) : null
            }
          ]}
        />
      </Card>
      <Card title="配置变更日志">
        <Table
          rowKey="id"
          dataSource={logs}
          pagination={false}
          columns={[
            { title: '配置键', dataIndex: 'key' },
            { title: '动作', dataIndex: 'action', render: (action) => <Tag color={action === 'rollback' ? 'orange' : 'blue'}>{action}</Tag> },
            { title: '旧值', dataIndex: 'old_value' },
            { title: '新值', dataIndex: 'new_value' },
            { title: '版本', dataIndex: 'version' },
            { title: '时间', dataIndex: 'created_at' }
          ]}
        />
      </Card>
    </Space>
  );
}

function AuthPage() {
  const [tokenInfo, setTokenInfo] = useState<TokenResponse | null>(null);

  const saveToken = (resp: TokenResponse) => {
    localStorage.setItem('eduadcrm.access_token', resp.access_token);
    localStorage.setItem('eduadcrm.refresh_token', resp.refresh_token);
    setTokenInfo(resp);
  };

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <Typography.Title level={3}>用户鉴权</Typography.Title>
      <Tabs
        items={[
          {
            key: 'register',
            label: '注册租户',
            children: (
              <Card>
                <Form layout="vertical" onFinish={async (values) => {
                  const resp = await api<TokenResponse>('/api/auth/register', { method: 'POST', body: JSON.stringify(values) });
                  saveToken(resp);
                  message.success('注册成功');
                }}>
                  <Form.Item name="tenant_name" label="机构名称" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item name="email" label="邮箱" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item name="display_name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true, min: 8 }]}><Input.Password /></Form.Item>
                  <Button type="primary" htmlType="submit">注册并登录</Button>
                </Form>
              </Card>
            )
          },
          {
            key: 'login',
            label: '登录',
            children: (
              <Card>
                <Form layout="vertical" onFinish={async (values) => {
                  const resp = await api<TokenResponse>('/api/auth/login', { method: 'POST', body: JSON.stringify(values) });
                  saveToken(resp);
                  message.success('登录成功');
                }}>
                  <Form.Item name="email" label="邮箱" rules={[{ required: true }]}><Input /></Form.Item>
                  <Form.Item name="password" label="密码" rules={[{ required: true }]}><Input.Password /></Form.Item>
                  <Button type="primary" htmlType="submit">登录</Button>
                </Form>
              </Card>
            )
          }
        ]}
      />
      {tokenInfo && (
        <Card title="当前令牌">
          <Space direction="vertical">
            <Typography.Text>Access Token 过期：{tokenInfo.access_token_expires_at}</Typography.Text>
            <Typography.Text>Refresh Token 过期：{tokenInfo.refresh_token_expires_at}</Typography.Text>
            <Typography.Text>用户：{tokenInfo.user.display_name} / {tokenInfo.user.email}</Typography.Text>
          </Space>
        </Card>
      )}
    </Space>
  );
}

function OAuthPage() {
  const [platform, setPlatform] = useState('douyin');
  const [authUrl, setAuthUrl] = useState('');

  return (
    <Space direction="vertical" size={16} className="page-stack">
      <Typography.Title level={3}>广告授权</Typography.Title>
      <Card>
        <Space>
          <Select
            value={platform}
            onChange={setPlatform}
            options={[
              { value: 'douyin', label: '抖音/巨量引擎' },
              { value: 'tencent', label: '腾讯广告' },
              { value: 'baidu', label: '百度营销' },
              { value: 'xiaohongshu', label: '小红书聚光' }
            ]}
          />
          <Button type="primary" onClick={async () => {
            const resp = await api<{ auth_url: string }>(`/api/oauth/${platform}/authorize`, {
              method: 'POST',
              body: JSON.stringify({
                tenant_id: 'demo-tenant',
                user_id: 'demo-user',
                organization_id: 'demo-org',
                channel_id: `channel-${platform}`,
                redirect_after_success: window.location.href
              })
            });
            setAuthUrl(resp.auth_url);
          }}>生成授权链接</Button>
        </Space>
      </Card>
      {authUrl && (
        <Card title="授权链接">
          <Typography.Paragraph copyable>{authUrl}</Typography.Paragraph>
        </Card>
      )}
    </Space>
  );
}

function OverviewPage() {
  return (
    <Card>
      <Typography.Title level={3}>平台概览</Typography.Title>
      <Typography.Paragraph>
        当前已落地 monorepo、配置服务、自有鉴权、广告 OAuth 回调骨架和 Python 数据面服务入口。
      </Typography.Paragraph>
    </Card>
  );
}

createRoot(document.getElementById('root')!).render(<App />);
