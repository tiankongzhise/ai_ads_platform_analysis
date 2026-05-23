# EduAdCRM UI 设计系统

> **项目**：EduAdCRM-MVP（教育行业广告×CRM数据平台）
> **版本**：v1.0
> **更新日期**：2026-04-27
> **技术栈**：React 18 + Ant Design Pro 6 + ECharts + Tailwind CSS

---

## 🎨 设计理念

**定位**：专业、高效、可信赖的 B 端数据管理平台
**关键词**：数据驱动 · 清晰层级 · 操作高效 · 教育行业亲和

### 设计原则

1. **数据优先**：仪表盘和信息架构以数据呈现为核心，图表清晰易读
2. **层级分明**：通过色彩、间距、阴影建立清晰的视觉层次
3. **操作高效**：常用操作路径短，减少点击次数
4. **行业亲和**：色彩搭配温和专业，符合教育行业审美偏好

---

## 🎨 色彩系统

### 主色调（Primary Colors）

```css
/* 品牌主色 - 深蓝色系，传达专业与信任 */
--color-primary-50: #EEF2FF;   /* 极浅蓝，用于背景 */
--color-primary-100: #E0E7FF;  /* 浅蓝，hover 背景 */
--color-primary-200: #C7D2FE;  /* 边框、分割线 */
--color-primary-300: #A5B4FC;  /* 次要强调 */
--color-primary-400: #818CF8;  /* 图标、标签 */
--color-primary-500: #6366F1;  /* 主按钮、链接 */
--color-primary-600: #4F46E5;  /* 主按钮 hover */
--color-primary-700: #4338CA;  /* 深色强调 */
--color-primary-800: #3730A3;  /* 深色文字 */
--color-primary-900: #312E81;  /* 最深色 */
```

### 辅助色（Secondary Colors）

```css
/* 中性色 - 灰蓝色系，用于文字和背景 */
--color-secondary-25: #FCFCFD; /* 页面背景 */
--color-secondary-50: #F9FAFB; /* 卡片背景 */
--color-secondary-100: #F3F4F6;/* 分割线 */
--color-secondary-200: #E5E7EB;/* 禁用边框 */
--color-secondary-300: #D1D5DB;/* 占位符 */
--color-secondary-400: #9CA3AF;/* 次要文字 */
--color-secondary-500: #6B7280;/* 次要文字 */
--color-secondary-600: #4B5563;/* 正文文字 */
--color-secondary-700: #374151;/* 标题文字 */
--color-secondary-800: #1F2937;/* 深色文字 */
--color-secondary-900: #111827;/* 最深文字 */

/* 功能色 - 语义色 */
--color-success: #10B981;      /* 成功、增长、正向指标 */
--color-success-light: #D1FAE5;
--color-warning: #F59E0B;      /* 警告、待处理 */
--color-warning-light: #FEF3C7;
--color-error: #EF4444;        /* 错误、负向指标、删除 */
--color-error-light: #FEE2E2;
--color-info: #3B82F6;         /* 信息、提示 */
--color-info-light: #DBEAFE;
```

### 广告平台品牌色

```css
/* 巨量引擎（抖音） */
--color-juliang: #FE2C55;      /* 抖音红 */
--color-juliang-light: #FFF0F2;

/* 百度营销 */
--color-baidu: #2932E1;         /* 百度蓝 */
--color-baidu-light: #EEF0FF;

/* 微信小程序 */
--color-wechat: #07C160;       /* 微信绿 */
--color-wechat-light: #E8FEF0;
```

### 图表配色方案

```css
/* 图表系列色 - 最多6个系列 */
--chart-series-1: #6366F1;     /* 主系列 - 品牌蓝 */
--chart-series-2: #10B981;     /* 次系列 - 成功绿 */
--chart-series-3: #F59E0B;     /* 第三系列 - 警告黄 */
--chart-series-4: #EF4444;     /* 第四系列 - 错误红 */
--chart-series-5: #8B5CF6;     /* 第五系列 - 紫色 */
--chart-series-6: #06B6D4;     /* 第六系列 - 青色 */
```

### 暗色模式支持

```css
[data-theme="dark"] {
  --color-bg-primary: #0F172A;     /* 页面背景 */
  --color-bg-secondary: #1E293B;   /* 卡片背景 */
  --color-bg-tertiary: #334155;    /* 表格斑马纹 */
  --color-text-primary: #F8FAFC;   /* 主文字 */
  --color-text-secondary: #94A3B8; /* 次要文字 */
  --color-border: #475569;         /* 边框 */
}
```

---

## 📝 字体系统

### 字体家族

```css
/* 主字体 - 中文优先，系统回退 */
--font-family-primary: 'Inter', 'PingFang SC', 'Microsoft YaHei', -apple-system, sans-serif;

/* 数字字体 - 图表和数据展示 */
--font-family-number: 'DIN Alternate', 'Roboto Mono', 'JetBrains Mono', monospace;

/* 代码字体 */
--font-family-mono: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
```

### 字体层级

```css
/* 标题系统 */
--font-size-display: 2.25rem;    /* 36px - 页面大标题 */
--font-size-h1: 1.875rem;        /* 30px - 模块标题 */
--font-size-h2: 1.5rem;          /* 24px - 区块标题 */
--font-size-h3: 1.25rem;         /* 20px - 卡片标题 */
--font-size-h4: 1.125rem;        /* 18px - 子标题 */

/* 正文系统 */
--font-size-base: 1rem;          /* 16px - 正文 */
--font-size-sm: 0.875rem;        /* 14px - 辅助文字 */
--font-size-xs: 0.75rem;         /* 12px - 标签、徽章 */

/* 行高 */
--line-height-tight: 1.25;       /* 标题 */
--line-height-normal: 1.5;       /* 正文 */
--line-height-relaxed: 1.75;     /* 长文本 */

/* 字重 */
--font-weight-normal: 400;
--font-weight-medium: 500;
--font-weight-semibold: 600;
--font-weight-bold: 700;
```

### 数字展示规范

```css
/* 数据指标数字 - 使用等宽字体 */
.data-number {
  font-family: var(--font-family-number);
  font-variant-numeric: tabular-nums;  /* 数字等宽，便于竖向对齐 */
  letter-spacing: -0.02em;
}

/* 数字大小规范 */
.data-display { font-size: 2.25rem; font-weight: 700; }  /* 核心指标 */
.data-large { font-size: 1.5rem; font-weight: 600; }     /* 次级指标 */
.data-normal { font-size: 1rem; font-weight: 500; }      /* 表格数据 */
```

---

## 📏 间距系统

### 基础间距（8px 基准）

```css
--space-0: 0;
--space-1: 0.25rem;    /* 4px - 紧凑间距 */
--space-2: 0.5rem;     /* 8px - 元素内间距 */
--space-3: 0.75rem;    /* 12px - 小间距 */
--space-4: 1rem;       /* 16px - 标准间距 */
--space-5: 1.25rem;    /* 20px - 中等间距 */
--space-6: 1.5rem;     /* 24px - 区块间距 */
--space-8: 2rem;       /* 32px - 大间距 */
--space-10: 2.5rem;    /* 40px - 页面内间距 */
--space-12: 3rem;      /* 48px - 区域间距 */
--space-16: 4rem;      /* 64px - 页面间距 */
```

### 组件间距规范

```css
/* 卡片内间距 */
.card-padding-sm { padding: var(--space-4); }   /* 小卡片 */
.card-padding-md { padding: var(--space-6); }   /* 标准卡片 */
.card-padding-lg { padding: var(--space-8); }   /* 大卡片 */

/* 表格间距 */
.table-cell-padding: var(--space-3) var(--space-4);
.table-row-gap: var(--space-1);

/* 表单间距 */
.form-item-gap: var(--space-6);
.form-label-margin: var(--space-2);
```

### 页面布局间距

```css
/* 页面容器 */
.page-padding: var(--space-6);
.page-max-width: 1440px;

/* 栅格间距 */
.grid-gutter: var(--space-6);

/* 侧边栏宽度 */
.sidebar-width-collapsed: 80px;
sidebar-width-expanded: 240px;
sidebar-width-submenu: 200px;
```

---

## 🎭 阴影系统

### 阴影层级

```css
/* 微阴影 - 卡片悬停 */
--shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);

/* 标准阴影 - 卡片默认 */
--shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);

/* 高阴影 - 弹窗、Dropdown */
--shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);

/* 最高阴影 - 模态框 */
--shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);

/* 聚焦阴影 - 输入框聚焦 */
--shadow-focus: 0 0 0 3px rgb(99 102 241 / 0.15);

/* 彩色阴影 - 品牌色 */
--shadow-primary: 0 4px 14px 0 rgb(99 102 241 / 0.35);
```

---

## 🔲 圆角系统

```css
/* 圆角规范 */
--radius-none: 0;
--radius-sm: 0.25rem;    /* 4px - 小按钮、标签 */
--radius-md: 0.375rem;   /* 6px - 输入框、小卡片 */
--radius-lg: 0.5rem;     /* 8px - 按钮、卡片 */
--radius-xl: 0.75rem;    /* 12px - 大卡片、模态框 */
--radius-2xl: 1rem;      /* 16px - 大容器 */
--radius-full: 9999px;   /* 圆形按钮、头像 */
```

---

## ⏱️ 动效系统

### 过渡时长

```css
/* 快速动画 - 微交互 */
--transition-fast: 150ms ease;

/* 标准动画 - 状态变化 */
--transition-normal: 250ms ease;

/* 慢速动画 - 页面切换 */
--transition-slow: 350ms ease;

/* 弹性动画 */
--transition-bounce: 300ms cubic-bezier(0.34, 1.56, 0.64, 1);
```

### 动画使用场景

| 场景 | 动画类型 | 时长 | 缓动函数 |
|------|----------|------|----------|
| 按钮 hover | Transform + Color | 150ms | ease |
| 卡片 hover | Transform + Shadow | 250ms | ease |
| 展开/收起 | Height + Opacity | 250ms | ease-in-out |
| 页面切换 | Fade + Slide | 300ms | ease-out |
| 模态框 | Scale + Fade | 250ms | ease-out |
| Toast | Slide + Fade | 300ms | ease |

### 动画使用原则

```css
/* 减少动画（尊重用户偏好）*/
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 🧱 组件库

### 1. 指标卡片 (MetricCard)

**用途**：展示核心业务指标（花费、线索、成本等）

**尺寸规范**：
- 高度：120px
- 内边距：24px
- 图标区域：48px × 48px

**视觉规范**：
```css
.metric-card {
  background: white;
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-secondary-100);
  transition: all var(--transition-normal);
}

.metric-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}
```

**状态**：
- 默认：白底、浅阴影
- 悬停：阴影加深、上移 2px
- 加载中：骨架屏动画
- 空数据：显示 "--" 或 "暂无数据"

**数据展示**：
```
┌────────────────────────────────────────────┐
│  📊 今日花费                    ↑ 12.5%    │
│                                            │
│  ¥ 12,580.00                               │
│                                            │
│  较昨日 +1,420.00                         │
└────────────────────────────────────────────┘
```

### 2. 图表容器 (ChartCard)

**用途**：封装 ECharts 图表，统一容器样式

**规范**：
```css
.chart-card {
  background: white;
  border-radius: var(--radius-xl);
  padding: var(--space-6);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-secondary-100);
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);
}

.chart-title {
  font-size: var(--font-size-h4);
  font-weight: var(--font-weight-semibold);
  color: var(--color-secondary-800);
}
```

**图表配色**（ECharts 主题）：
```javascript
const chartTheme = {
  color: ['#6366F1', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4'],
  // 文字颜色
  textStyle: { color: '#6B7280' },
  // 坐标轴
  xAxis: { axisLine: { lineStyle: { color: '#E5E7EB' } } },
  yAxis: { axisLine: { lineStyle: { color: '#E5E7EB' } } },
  // 网格
  grid: { borderColor: '#F3F4F6' },
};
```

### 3. 数据表格 (DataTable)

**基于 Ant Design Table 定制**：

```css
/* 表头 */
.table-header {
  background: var(--color-secondary-50);
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
  color: var(--color-secondary-600);
}

/* 斑马纹 */
.table-row-even { background: white; }
.table-row-odd { background: var(--color-secondary-50); }

/* 悬停行 */
.table-row:hover { background: var(--color-primary-50); }
```

**操作列规范**：
- 图标按钮：24px × 24px
- 间距：8px
- 悬停提示：Tooltip

### 4. 表单组件 (FormComponents)

**输入框规范**：
```css
.input {
  height: 40px;
  padding: 0 var(--space-3);
  border: 1px solid var(--color-secondary-200);
  border-radius: var(--radius-md);
  font-size: var(--font-size-base);
  transition: all var(--transition-fast);
}

.input:focus {
  border-color: var(--color-primary-500);
  box-shadow: var(--shadow-focus);
  outline: none;
}

.input::placeholder {
  color: var(--color-secondary-300);
}

.input:disabled {
  background: var(--color-secondary-50);
  cursor: not-allowed;
}
```

**按钮规范**：
```css
/* 主要按钮 */
.btn-primary {
  height: 40px;
  padding: 0 var(--space-6);
  background: var(--color-primary-500);
  color: white;
  border-radius: var(--radius-md);
  font-weight: var(--font-weight-medium);
  transition: all var(--transition-fast);
}

.btn-primary:hover {
  background: var(--color-primary-600);
  box-shadow: var(--shadow-primary);
}

/* 次要按钮 */
.btn-secondary {
  height: 40px;
  padding: 0 var(--space-6);
  background: white;
  color: var(--color-secondary-700);
  border: 1px solid var(--color-secondary-200);
  border-radius: var(--radius-md);
}

.btn-secondary:hover {
  border-color: var(--color-primary-500);
  color: var(--color-primary-500);
}

/* 危险按钮 */
.btn-danger {
  background: var(--color-error);
  color: white;
}

.btn-danger:hover {
  background: #DC2626;
}
```

### 5. 状态徽章 (StatusBadge)

```css
.badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

/* 状态变体 */
.badge-success { background: var(--color-success-light); color: #059669; }
.badge-warning { background: var(--color-warning-light); color: #D97706; }
.badge-error { background: var(--color-error-light); color: #DC2626; }
.badge-info { background: var(--color-info-light); color: #2563EB; }
.badge-default { background: var(--color-secondary-100); color: var(--color-secondary-600); }
```

### 6. 侧边导航 (SiderMenu)

```css
/* 侧边栏 */
.sider {
  width: 240px;
  background: white;
  border-right: 1px solid var(--color-secondary-100);
}

/* 菜单项 */
.menu-item {
  height: 44px;
  padding: 0 var(--space-4);
  margin: 4px 8px;
  border-radius: var(--radius-md);
  color: var(--color-secondary-600);
  transition: all var(--transition-fast);
}

.menu-item:hover {
  background: var(--color-secondary-50);
  color: var(--color-secondary-800);
}

.menu-item-active {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
  font-weight: var(--font-weight-medium);
}

/* 图标 */
.menu-icon {
  width: 20px;
  height: 20px;
  margin-right: var(--space-3);
}
```

---

## 📱 页面布局

### 整体布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│ Header（64px）                                                  │
│ [Logo] [搜索框] [通知] [用户菜单]                               │
├────────┬────────────────────────────────────────────────────────┤
│ Sider  │ Content Area                                          │
│ (240px)│                                                         │
│        │ ┌──────────────────────────────────────────────────┐  │
│ 导航   │ │ Page Header                                       │  │
│        │ │ [页面标题] [面包屑] [操作按钮]                    │  │
│        │ └──────────────────────────────────────────────────┘  │
│        │                                                         │
│        │ ┌──────────────────────────────────────────────────┐  │
│        │ │ Content                                          │  │
│        │ │                                                  │  │
│        │ │                                                  │  │
│        │ └──────────────────────────────────────────────────┘  │
└────────┴────────────────────────────────────────────────────────┘
```

### 主要页面设计

#### 1. 登录页 (Login)

**设计要点**：
- 简洁大气，品牌色突出
- 表单居中，视觉焦点集中
- 支持记住密码、忘记密码

**布局**：
```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│    ┌────────────────────────────────────────────────────┐   │
│    │                                                    │   │
│    │              🎯 EduAdCRM                          │   │
│    │           教育广告数据管理平台                      │   │
│    │                                                    │   │
│    │  ┌────────────────────────────────────────────┐   │   │
│    │  │  邮箱地址                                  │   │   │
│    │  └────────────────────────────────────────────┘   │   │
│    │                                                    │   │
│    │  ┌────────────────────────────────────────────┐   │   │
│    │  │  密码                                       │   │   │
│    │  └────────────────────────────────────────────┘   │   │
│    │                                                    │   │
│    │  ☑ 记住密码                                       │   │
│    │                                                    │   │
│    │  ┌────────────────────────────────────────────┐   │   │
│    │  │              登 录                          │   │   │
│    │  └────────────────────────────────────────────┘   │   │
│    │                                                    │   │
│    │          还没有账号？立即注册 →                   │   │
│    └────────────────────────────────────────────────────┘   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

#### 2. ROI 仪表盘 (Dashboard)

**核心页面，展示关键业务指标**

**布局分区**：
```
┌─────────────────────────────────────────────────────────────────┐
│  ROI 仪表盘                                           [日期选择]│
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │ 今日花费 │ │ 累计线索 │ │ 线索成本 │ │   ROI   │ │ 转化率  │   │
│  │ ¥12,580 │ │   328   │ │  ¥38.4  │ │  1:4.2  │ │  2.8%   │   │
│  │ ↑12.5% │ │ ↑8.3%  │ │ ↓5.2%  │ │ ↑0.3   │ │ ↑0.2%  │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
│                                                                 │
│  ┌───────────────────────────────────┐ ┌───────────────────────┐│
│  │                                   │ │                       ││
│  │       花费与线索趋势图             │ │    渠道对比           ││
│  │       （近30天折线图）            │ │    （柱状图）         ││
│  │                                   │ │                       ││
│  │                                   │ │                       ││
│  │                                   │ │                       ││
│  └───────────────────────────────────┘ └───────────────────────┘│
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                    广告计划效果排行                         │ │
│  │  #  │ 计划名称   │ 花费   │ 线索  │ 成本  │ 转化率 │ 操作  │ │
│  │  1  │ 考研课程A  │ ¥3,200 │  86   │ ¥37.2 │ 3.1%   │ 查看  │ │
│  │  2  │ 托福冲刺B  │ ¥2,850 │  72   │ ¥39.6 │ 2.8%   │ 查看  │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

**图表设计规范**：

1. **趋势折线图**：
   - 双 Y 轴：左侧花费（柱状）、右侧线索数（折线）
   - 悬停显示详情 tooltip
   - 支持时间范围选择

2. **渠道对比图**：
   - 水平柱状图
   - 显示抖音 vs 百度的花费和线索对比
   - 使用平台品牌色

3. **数据表格**：
   - 支持排序
   - 分页展示
   - 行悬停高亮

#### 3. CRM 线索管理 (CRM Leads)

**功能**：Excel 导入、线索列表、状态管理

**布局**：
```
┌─────────────────────────────────────────────────────────────────┐
│ CRM 线索管理                                                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────┐  ┌─────────────────────────────────┐   │
│  │                     │  │  ┌─────────────────────────┐   │   │
│  │   📤 上传 Excel      │  │  │ 姓名/电话: [__________]  │   │   │
│  │                     │  │  │ 来源渠道: [全部 ▼]       │   │   │
│  │  ┌───────────────┐  │  │  │ 成交状态: [全部 ▼]       │   │   │
│  │  │   拖拽文件    │  │  │  │ 导入时间: [____至____]   │   │   │
│  │  │   到此处      │  │  │  └─────────────────────────┘   │   │
│  │  │               │  │  │                                  │   │
│  │  └───────────────┘  │  │  [重置]  [搜索]                  │   │
│  │                     │  │                                   │   │
│  │  支持 .xlsx, .csv   │  └─────────────────────────────────┘   │
│  │  最大 10MB          │                                       │
│  └─────────────────────┘                                       │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │  # │ 姓名    │ 电话(脱敏) │ 来源  │ 课程    │ 状态  │ 操作 │ │
│  │  1 │ 张三    │ 138****1234│ 抖音  │ 考研课程│ 已成交│ ⋮   │ │
│  │  2 │ 李四    │ 139****5678│ 百度  │ 托福    │ 跟进中│ ⋮   │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
│  共 328 条记录                        < 1 2 3 ... 33 >  [20条/页]│
└─────────────────────────────────────────────────────────────────┘
```

**上传流程状态**：
```
┌─────────────────────────────────────────────────────────────┐
│  导入进度                                              [关闭]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│       ████████████████████░░░░░░░░░  65%                   │
│                                                             │
│  已处理: 32500 / 50000 条                                    │
│  成功: 31,200 条 ✓                                          │
│  跳过: 800 条 (重复)                                         │
│  失败: 500 条 (格式错误) ⚠                                   │
│                                                             │
│                          [取消导入]  [下载错误报告]          │
└─────────────────────────────────────────────────────────────┘
```

#### 4. 广告账户管理 (Ad Accounts)

**功能**：绑定巨量引擎、百度广告账户，数据同步状态

**布局**：
```
┌─────────────────────────────────────────────────────────────────┐
│ 广告账户管理                                                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 抖音广告账户                              [+ 绑定新账户] │   │
│  ├─────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────────────────────────────────────────────┐│   │
│  │  │ 🔴 抖音 │ 抖音教育官方账户                            ││   │
│  │  │         │ 账户ID: 1234567890                         ││   │
│  │  │         │ 余额: ¥5,280.00 | 状态: 活跃               ││   │
│  │  │         │ 最后同步: 2026-04-27 16:30                 ││   │
│  │  │         │ [查看数据] [断开连接]                      ││   │
│  │  └─────────────────────────────────────────────────────┘│   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 百度营销账户                              [+ 绑定新账户] │   │
│  ├─────────────────────────────────────────────────────────┤   │
│  │  ┌─────────────────────────────────────────────────────┐│   │
│  │  │ 🔵 百度 │ 百度留学推广账户                           ││   │
│  │  │         │ 账户ID: 9876543210                        ││   │
│  │  │         │ 余额: ¥3,150.00 | 状态: 活跃              ││   │
│  │  │         │ 最后同步: 2026-04-27 15:45               ││   │
│  │  │         │ [查看数据] [同步数据] [断开连接]          ││   │
│  │  └─────────────────────────────────────────────────────┘│   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

#### 5. 交叉分析 (Analytics)

**功能**：广告计划×课程交叉分析

**布局**：
```
┌─────────────────────────────────────────────────────────────────┐
│ 交叉分析                                              [导出CSV]│
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ [时间范围选择]  [广告平台 ▼]  [课程类型 ▼]              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │           │ 考研课程  │ 托福课程 │ 雅思课程  │ 合计       │ │
│  │ ─────────────────────────────────────────────────────────│ │
│  │ 抖音-品牌 │  ¥12,500  │  ¥8,200  │  ¥6,800   │  ¥27,500  │ │
│  │ 抖音-效果 │  ¥8,300   │  ¥5,600  │  ¥4,200   │  ¥18,100  │ │
│  │ 百度-搜索 │  ¥6,800   │  ¥9,200  │  ¥5,400   │  ¥21,400  │ │
│  │ 百度-信息流│  ¥4,200   │  ¥3,800  │  ¥2,600   │  ¥10,600  │ │
│  │ ─────────────────────────────────────────────────────────│ │
│  │ 合计     │  ¥31,800  │  ¥26,800 │  ¥19,000  │  ¥77,600  │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                    交叉分析热力图                          │ │
│  │                                                           │ │
│  │      考研   托福   雅思                                    │ │
│  │  抖音 ████  ███   ██     ← 考研线索最多                   │ │
│  │  百度 ███   ████  ███                                      │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## ♿ 可访问性标准

### WCAG AA 合规

```css
/* 色彩对比度检查 */
/* 正文文字：最小 4.5:1 */
/* 大文字（18px+）：最小 3:1 */
/* 图表标注：最小 3:1 */

/* 示例对比色组合 */
.contrast-pass-1 { color: #111827; background: #FFFFFF; }      /* 15.8:1 ✓ */
.contrast-pass-2 { color: #6366F1; background: #FFFFFF; }       /* 4.6:1 ✓ */
.contrast-pass-3 { color: #10B981; background: #FFFFFF; }       /* 3.1:1 ✓ */
```

### 键盘导航

```css
/* 焦点样式 */
:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

/* 焦点顺序逻辑 */
.tab-order {
  tab-order: logical;  /* 符合视觉顺序 */
}

/* 跳转链接 */
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  padding: 8px 16px;
  background: var(--color-primary-500);
  color: white;
  z-index: 1000;
}

.skip-link:focus {
  top: 0;
}
```

### ARIA 标签

```html
<!-- 图表替代文字 -->
<div role="img" aria-label="近30天花费与线索趋势图">
  <!-- ECharts 图表 -->
</div>

<!-- 数据表格 -->
<table role="grid" aria-label="广告计划效果数据">
  <thead>
    <tr>
      <th scope="col" aria-sort="descending">花费</th>
    </tr>
  </thead>
</table>

<!-- 加载状态 -->
<div role="status" aria-live="polite">
  加载中，请稍候...
</div>
```

### 触摸目标

```css
/* 移动端触摸目标最小 44px × 44px */
@media (max-width: 768px) {
  .touch-target {
    min-width: 44px;
    min-height: 44px;
  }

  .btn {
    min-height: 44px;
    padding: 0 var(--space-6);
  }

  .menu-item {
    min-height: 44px;
  }
}
```

---

## 📐 响应式设计

### 断点策略

```css
/* 移动端优先 */

/* 基础（移动端）320px - 639px */
.container { max-width: 100%; padding: var(--space-4); }

/* 小型平板 640px+ */
@media (min-width: 640px) {
  .container { max-width: 640px; }
  .grid-cols-2 { grid-template-columns: repeat(2, 1fr); }
}

/* 平板 768px+ */
@media (min-width: 768px) {
  .container { max-width: 768px; }
  .grid-cols-3 { grid-template-columns: repeat(3, 1fr); }
}

/* 小桌面 1024px+ */
@media (min-width: 1024px) {
  .container { max-width: 1024px; }
  .layout-with-sider { display: flex; }
  .sider { width: 240px; }
  .content { flex: 1; }
}

/* 桌面 1280px+ */
@media (min-width: 1280px) {
  .container { max-width: 1280px; }
  .grid-cols-4 { grid-template-columns: repeat(4, 1fr); }
}

/* 大桌面 1536px+ */
@media (min-width: 1536px) {
  .container { max-width: 1440px; }
}
```

### 响应式组件策略

| 组件 | 桌面 (≥1024px) | 平板 (768-1023px) | 移动端 (<768px) |
|------|----------------|-------------------|-----------------|
| 侧边导航 | 固定显示 | 可折叠 | 底部 TabBar |
| 指标卡片 | 5列网格 | 3列网格 | 1列堆叠 |
| 图表 | 并排双图 | 上下堆叠 | 简化单图 |
| 表格 | 全字段 | 隐藏次要字段 | 卡片化展示 |
| 表单 | 横向排列 | 纵向排列 | 纵向堆叠 |

---

## 📋 组件文档示例

### MetricCard 使用文档

```tsx
import { MetricCard } from '@/components/MetricCard';
import { TrendingUp, TrendingDown } from '@ant-design/icons';

// 基本用法
<MetricCard
  title="今日花费"
  value={12580.00}
  prefix="¥"
  precision={2}
  trend={{ value: 12.5, direction: 'up' }}
  trendValue="较昨日 +1,420"
/>

// 属性说明
interface MetricCardProps {
  title: string;           // 指标名称
  value: number | string;  // 数值
  prefix?: string;         // 前缀符号（如 ¥, %, :）
  suffix?: string;         // 后缀符号
  precision?: number;      // 小数位数
  trend?: {
    value: number;         // 趋势百分比
    direction: 'up' | 'down' | 'flat';
  };
  trendValue?: string;     // 趋势说明
  icon?: ReactNode;        // 图标
  loading?: boolean;       // 加载状态
  empty?: string;          // 空数据文案
}
```

### ChartCard 使用文档

```tsx
import { ChartCard } from '@/components/ChartCard';
import { TrendChart } from '@/components/charts/TrendChart';
import { Button, Space } from 'antd';

// 用法
<ChartCard
  title="花费与线索趋势"
  extra={
    <Space>
      <Button type="text" size="small">7天</Button>
      <Button type="primary" size="small">30天</Button>
      <Button type="text" size="small">90天</Button>
    </Space>
  }
>
  <TrendChart data={trendData} />
</ChartCard>

// 属性说明
interface ChartCardProps {
  title: string;           // 图表标题
  extra?: ReactNode;       // 标题右侧操作区
  loading?: boolean;       // 加载状态
  children: ReactNode;     // 图表组件
  height?: number;         // 图表高度，默认 300
}
```

---

## 🎨 图标系统

### 图标选择原则

1. **使用 Ant Design Icons**（与 Ant Design 保持一致）
2. **语义明确**：每个操作/状态有唯一图标
3. **统一尺寸**：标准 20px，特殊场景 16px/24px

### 核心图标映射

| 场景 | 图标 | Ant Design 名称 |
|------|------|-----------------|
| 仪表盘 | 📊 | DashboardOutlined |
| CRM/线索 | 👥 | TeamOutlined |
| 广告账户 | 📢 | notificationOutlined |
| 数据分析 | 📈 | AnalysisOutlined |
| 报表 | 📋 | FileTextOutlined |
| 设置 | ⚙️ | SettingOutlined |
| 上传 | 📤 | UploadOutlined |
| 下载 | 📥 | DownloadOutlined |
| 编辑 | ✏️ | EditOutlined |
| 删除 | 🗑️ | DeleteOutlined |
| 查看 | 👁️ | EyeOutlined |
| 成功 | ✓ | CheckCircleOutlined |
| 警告 | ⚠️ | WarningOutlined |
| 错误 | ✕ | CloseCircleOutlined |
| 同步 | 🔄 | SyncOutlined |

---

## 📄 状态设计

### 加载状态

```css
/* 骨架屏 */
.skeleton {
  background: linear-gradient(
    90deg,
    var(--color-secondary-100) 25%,
    var(--color-secondary-50) 50%,
    var(--color-secondary-100) 75%
  );
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
}

@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* 加载动画（小型） */
.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--color-secondary-200);
  border-top-color: var(--color-primary-500);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
```

### 空状态

```tsx
// 空数据展示组件
const EmptyState = ({ type, action }) => (
  <div className="empty-state">
    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={
      type === 'no-data'
        ? '暂无数据'
        : type === 'no-results'
        ? '没有找到匹配的记录'
        : '数据加载失败'
    }>
      {action && <Button type="primary" onClick={action.onClick}>
        {action.label}
      </Button>}
    </Empty>
  </div>
);
```

### 错误状态

```tsx
// 错误边界组件
const ErrorState = ({ error, onRetry }) => (
  <Result
    status="error"
    title="加载失败"
    subTitle={error.message || '请稍后重试'}
    extra={<Button onClick={onRetry}>重新加载</Button>}
  />
);
```

---

## 🎯 设计验收清单

### 视觉验收

- [ ] 所有页面符合设计规范（颜色、字体、间距）
- [ ] 组件状态完整（默认、悬停、激活、禁用、加载）
- [ ] 图表配色符合品牌规范
- [ ] 响应式布局在各断点表现正常
- [ ] 深色模式切换正常

### 功能验收

- [ ] 表单验证提示清晰
- [ ] 操作反馈及时（成功/失败 Toast）
- [ ] 加载状态优雅（骨架屏/Loading）
- [ ] 空数据状态友好
- [ ] 错误处理完善

### 可访问性验收

- [ ] 色彩对比度符合 WCAG AA
- [ ] 键盘导航完整
- [ ] 屏幕阅读器语义正确
- [ ] 触摸目标符合要求

---

**文档版本**：v1.0
**维护者**：UI 设计团队
**更新日期**：2026-04-27
