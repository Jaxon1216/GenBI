import { listMyChartByPageUsingPost } from '@/services/yubi/chartController';
import SafeChart from '@/components/SafeChart';
import ChartCardSkeleton from '@/components/ChartCardSkeleton';
import VirtualDataTable from '@/components/VirtualDataTable';
import { usePolling } from '@/hooks/usePolling';
import { useModel } from '@@/exports';
import { DatabaseOutlined } from '@ant-design/icons';
import { Avatar, Button, Card, List, message, Result, Tag } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import Search from 'antd/es/input/Search';

/**
 * 我的图表页面
 * @constructor
 */
const MyChartPage: React.FC = () => {
  const initSearchParams = {
    // 默认第一页
    current: 1,
    // 每页展示4条数据
    pageSize: 4,
  };

  const [searchParams, setSearchParams] = useState<API.ChartQueryRequest>({ ...initSearchParams });
  // 从全局状态中获取到当前登录的用户信息
  const { initialState } = useModel('@@initialState');
  const { currentUser } = initialState ?? {};
  const [chartList, setChartList] = useState<API.Chart[]>();
  const [total, setTotal] = useState<number>(0);
  // 用来控制页面是否加载
  const [loading, setLoading] = useState<boolean>(true);
  const [dataModalChart, setDataModalChart] = useState<API.Chart | null>(null);

  const loadData = async (silent = false) => {
    if (!silent) setLoading(true);
    try {
      const res = await listMyChartByPageUsingPost(searchParams);
      if (res.data) {
        setChartList(res.data.records ?? []);
        setTotal(res.data.total ?? 0);
      } else if (!silent) {
        message.error('获取我的图表失败');
      }
    } catch (e: any) {
      if (!silent) message.error('获取我的图表失败，' + e.message);
    }
    if (!silent) setLoading(false);
  };

  const hasPendingCharts = useMemo(
    () => chartList?.some((c) => c.status === 'wait' || c.status === 'running') ?? false,
    [chartList],
  );

  const pollCallback = useCallback(async () => {
    await loadData(true);
  }, [searchParams]);

  const { reportDataSnapshot } = usePolling(pollCallback, {
    enabled: hasPendingCharts,
    baseInterval: 3000,
    maxInterval: 30000,
  });

  useEffect(() => {
    if (chartList) {
      const snapshot = chartList.map((c) => `${c.id}:${c.status}`).join(',');
      reportDataSnapshot(snapshot);
    }
  }, [chartList, reportDataSnapshot]);

  useEffect(() => {
    loadData();
  }, [searchParams]);

  return (
    <div className="my-chart-page">
      {/* 引入搜索框 */}
      <div>
        {/* 
          当用户点击搜索按钮触发 一定要把新设置的搜索条件初始化，要把页面切回到第一页;
          如果用户在第二页,输入了一个新的搜索关键词,应该重新展示第一页,而不是还在搜第二页的内容
        */}
        <Search placeholder="请输入图表名称" enterButton loading={loading} onSearch={(value) => {
          // 设置搜索条件
          setSearchParams({
            // 原始搜索条件
            ...initSearchParams,
            // 搜索词
            name: value,
          })
        }}/>
      </div>
      <div className="margin-16" />
      {loading && !chartList?.length ? (
        <ChartCardSkeleton count={4} />
      ) : (
      <List
        /*
          栅格间隔16像素;xs屏幕<576px,栅格数1;
          sm屏幕≥576px，栅格数1;md屏幕≥768px,栅格数1;
          lg屏幕≥992px,栅格数2;xl屏幕≥1200px,栅格数2;
          xxl屏幕≥1600px,栅格数2
        */
        grid={{
          gutter: 16,
          xs: 1,
          sm: 1,
          md: 1,
          lg: 2,
          xl: 2,
          xxl: 2,
        }}
        pagination={{
          /*
            page第几页，pageSize每页显示多少条;
            当用户点击这个分页组件,切换分页时,这个组件就会去触发onChange方法,会改变咱们现在这个页面的搜索条件
          */
          onChange: (page, pageSize) => {
            // 当切换分页，在当前搜索条件的基础上，把页数调整为当前的页数
            setSearchParams({
              ...searchParams,
              current: page,
              pageSize,
            })
          },
          // 显示当前页数
          current: searchParams.current,
          // 页面参数改成自己的
          pageSize: searchParams.pageSize,
          // 总数设置成自己的
          total: total,
        }}
        loading={loading}
        dataSource={chartList}
        renderItem={(item) => (
          <List.Item key={item.id}>
            <Card style={{ width: '100%' }}>
              <List.Item.Meta
                // 把当前登录用户信息的头像展示出来
                avatar={<Avatar src={currentUser && currentUser.userAvatar} />}
                title={item.name}
                description={item.chartType ? '图表类型：' + item.chartType : undefined}
              />
              <>
                {
                  item.status === 'wait' && <>
                    <Result
                      status="warning"
                      title="待生成"
                      subTitle={item.execMessage ?? '当前图表生成队列繁忙，请耐心等候'}
                      extra={<Tag color="orange">自动刷新中</Tag>}
                    />
                  </>
                }
                {
                  item.status === 'running' && <>
                    <Result
                      status="info"
                      title="图表生成中"
                      subTitle={item.execMessage}
                      extra={<Tag color="processing">自动刷新中</Tag>}
                    />
                  </>
                }
                {
                  item.status === 'succeed' && <>
                    <div style={{ marginBottom: 16 }} />
                    <p>{'分析目标：' + item.goal}</p>
                    <div style={{ marginBottom: 16 }} />
                    <SafeChart rawChartJson={item.genChart} stripTitle />
                    {item.chartData && (
                      <Button
                        type="link"
                        icon={<DatabaseOutlined />}
                        size="small"
                        style={{ marginTop: 8 }}
                        onClick={() => setDataModalChart(item)}
                      >
                        查看原始数据
                      </Button>
                    )}
                  </>
                }
                {
                  // 当状态（item.status）为'failed'时，显示生成失败的结果组件
                  item.status === 'failed' && <>
                    <Result
                      status="error"
                      title="图表生成失败"
                      subTitle={item.execMessage}
                    />
                  </>
                }
              </>
            </Card>
          </List.Item>
        )}
      />
      )}
      <VirtualDataTable
        visible={!!dataModalChart}
        onClose={() => setDataModalChart(null)}
        csvData={dataModalChart?.chartData}
        chartName={dataModalChart?.name}
      />
    </div>
  );
};
export default MyChartPage;
