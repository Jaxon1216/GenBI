import SafeChart from '@/components/SafeChart';
import { listMyChartByPageUsingPost } from '@/services/yubi/chartController';
import { AppstoreAddOutlined, DeleteOutlined, SaveOutlined, UndoOutlined } from '@ant-design/icons';
import { Button, Card, Empty, message, Modal, Select, Space, Spin, Tag } from 'antd';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import type { Layout, Layouts } from 'react-grid-layout';
import { Responsive, WidthProvider } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';

const ResponsiveGridLayout = WidthProvider(Responsive);

const STORAGE_KEY = 'genbi_dashboard';

interface DashboardItem {
  chartId: number;
  chart: API.Chart;
}

interface DashboardState {
  items: DashboardItem[];
  layouts: Layouts;
}

function loadDashboard(): DashboardState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {}
  return { items: [], layouts: {} };
}

function saveDashboard(state: DashboardState) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
}

const Dashboard: React.FC = () => {
  const [items, setItems] = useState<DashboardItem[]>([]);
  const [layouts, setLayouts] = useState<Layouts>({});
  const [addModalVisible, setAddModalVisible] = useState(false);
  const [availableCharts, setAvailableCharts] = useState<API.Chart[]>([]);
  const [loadingCharts, setLoadingCharts] = useState(false);
  const [selectedChartIds, setSelectedChartIds] = useState<number[]>([]);
  const isInitialized = useRef(false);

  useEffect(() => {
    const saved = loadDashboard();
    setItems(saved.items);
    setLayouts(saved.layouts);
    isInitialized.current = true;
  }, []);

  const handleSave = useCallback(() => {
    saveDashboard({ items, layouts });
    message.success('看板布局已保存');
  }, [items, layouts]);

  const handleLayoutChange = useCallback((_current: Layout[], allLayouts: Layouts) => {
    setLayouts(allLayouts);
  }, []);

  const handleRemoveItem = useCallback((chartId: number) => {
    setItems((prev) => prev.filter((i) => i.chartId !== chartId));
  }, []);

  const handleReset = useCallback(() => {
    Modal.confirm({
      title: '确认重置看板？',
      content: '将清除所有已添加的图表和布局',
      onOk: () => {
        setItems([]);
        setLayouts({});
        localStorage.removeItem(STORAGE_KEY);
        message.success('看板已重置');
      },
    });
  }, []);

  const openAddModal = useCallback(async () => {
    setAddModalVisible(true);
    setLoadingCharts(true);
    try {
      const allCharts: API.Chart[] = [];
      let current = 1;
      let hasMore = true;
      while (hasMore) {
        const res = await listMyChartByPageUsingPost({ current, pageSize: 20 });
        const records = res.data?.records ?? [];
        allCharts.push(...records);
        hasMore = records.length === 20;
        current++;
      }
      setAvailableCharts(allCharts.filter((c) => c.status === 'succeed'));
    } catch {
      message.error('获取图表列表失败');
    }
    setLoadingCharts(false);
  }, []);

  const handleAddCharts = useCallback(() => {
    const existingIds = new Set(items.map((i) => i.chartId));
    const newItems: DashboardItem[] = [];

    selectedChartIds.forEach((id) => {
      if (existingIds.has(id)) return;
      const chart = availableCharts.find((c) => c.id === id);
      if (chart) {
        newItems.push({ chartId: id, chart });
      }
    });

    if (newItems.length === 0) {
      message.warning('所选图表已在看板中');
    } else {
      setItems((prev) => [...prev, ...newItems]);
      message.success(`已添加 ${newItems.length} 个图表`);
    }

    setSelectedChartIds([]);
    setAddModalVisible(false);
  }, [selectedChartIds, availableCharts, items]);

  const generateDefaultLayout = useCallback(
    (allItems: DashboardItem[]): Layout[] => {
      return allItems.map((item, index) => {
        const existingLayout = layouts.lg?.find((l) => l.i === String(item.chartId));
        if (existingLayout) return existingLayout;
        return {
          i: String(item.chartId),
          x: (index % 2) * 6,
          y: Math.floor(index / 2) * 4,
          w: 6,
          h: 4,
          minW: 3,
          minH: 3,
        };
      });
    },
    [layouts],
  );

  return (
    <div>
      <Card
        title="数据看板"
        extra={
          <Space>
            <Button icon={<AppstoreAddOutlined />} type="primary" onClick={openAddModal}>
              添加图表
            </Button>
            <Button icon={<SaveOutlined />} onClick={handleSave}>
              保存布局
            </Button>
            <Button icon={<UndoOutlined />} danger onClick={handleReset}>
              重置
            </Button>
          </Space>
        }
      >
        {items.length === 0 ? (
          <Empty description="暂无图表，点击「添加图表」开始">
            <Button type="primary" onClick={openAddModal}>
              添加图表
            </Button>
          </Empty>
        ) : (
          <ResponsiveGridLayout
            className="layout"
            layouts={{ lg: generateDefaultLayout(items) }}
            breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
            cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
            rowHeight={80}
            onLayoutChange={handleLayoutChange}
            draggableHandle=".drag-handle"
            isResizable
            isDraggable
          >
            {items.map((item) => (
              <div key={String(item.chartId)}>
                <Card
                  size="small"
                  title={
                    <span className="drag-handle" style={{ cursor: 'grab' }}>
                      {item.chart.name || '未命名图表'}
                    </span>
                  }
                  extra={
                    <Space size="small">
                      {item.chart.chartType && <Tag color="blue">{item.chart.chartType}</Tag>}
                      <Button
                        type="text"
                        size="small"
                        danger
                        icon={<DeleteOutlined />}
                        onClick={() => handleRemoveItem(item.chartId)}
                      />
                    </Space>
                  }
                  style={{ height: '100%', overflow: 'hidden' }}
                  styles={{ body: { height: 'calc(100% - 40px)', padding: 8, overflow: 'hidden' } }}
                >
                  <SafeChart
                    rawChartJson={item.chart.genChart}
                    stripTitle
                    style={{ height: '100%', width: '100%' }}
                  />
                </Card>
              </div>
            ))}
          </ResponsiveGridLayout>
        )}
      </Card>

      <Modal
        title="添加图表到看板"
        open={addModalVisible}
        onOk={handleAddCharts}
        onCancel={() => {
          setAddModalVisible(false);
          setSelectedChartIds([]);
        }}
        okText="添加"
        cancelText="取消"
        okButtonProps={{ disabled: selectedChartIds.length === 0 }}
      >
        <Spin spinning={loadingCharts}>
          <Select
            mode="multiple"
            style={{ width: '100%' }}
            placeholder="选择要添加的图表"
            value={selectedChartIds}
            onChange={setSelectedChartIds}
            options={availableCharts.map((c) => ({
              label: `${c.name || '未命名'} (${c.chartType || '未知类型'})`,
              value: c.id!,
              disabled: items.some((i) => i.chartId === c.id),
            }))}
            optionFilterProp="label"
          />
        </Spin>
      </Modal>
    </div>
  );
};

export default Dashboard;
