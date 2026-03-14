import { listMyChartByPageUsingPost } from '@/services/yubi/chartController';
import { Card, Col, message, Pagination, Result, Row, Space, Spin, Tag, Typography } from 'antd';
import ReactECharts from 'echarts-for-react';
import React, { useEffect, useState } from 'react';
import './index.css';

const { Title, Paragraph } = Typography;

/**
 * 我的图表页面
 * @constructor
 */
const MyChartPage: React.FC = () => {
  const initSearchParams = {
    current: 1,
    pageSize: 4,
  };

  const [searchParams, setSearchParams] = useState<API.ChartQueryRequest>({
    ...initSearchParams,
  });
  const [chartList, setChartList] = useState<API.Chart[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(false);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await listMyChartByPageUsingPost(searchParams);

      if (res.data) {
        setChartList(res.data.records ?? []);
        setTotal(res.data.total ?? 0);
      } else {
        message.error('获取我的图表失败');
      }
    } catch (e: any) {
      message.error('获取我的图表失败,' + e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [searchParams]);

  const handlePageChange = (page: number, pageSize: number) => {
    setSearchParams({
      ...searchParams,
      current: page,
      pageSize,
    });
  };

  return (
    <div className="my-chart-page">
      <div className="page-header">
        <Title level={2}>我的图表</Title>
        <Paragraph type="secondary">查看和管理您生成的所有图表</Paragraph>
      </div>

      <Spin spinning={loading}>
        {chartList.length === 0 && !loading ? (
          <Result
            status="404"
            title="暂无图表"
            subTitle="您还没有创建任何图表，快去创建一个吧！"
          />
        ) : (
          <>
            <Row gutter={[16, 16]}>
              {chartList.map((chart) => {
                const chartOption = chart.genChart ? JSON.parse(chart.genChart) : null;
                return (
                  <Col xs={24} sm={24} md={12} lg={12} xl={12} key={chart.id}>
                    <Card
                      className="chart-card"
                      hoverable
                      title={
                        <Space>
                          <span>{chart.name || '未命名图表'}</span>
                          {chart.chartType && <Tag color="blue">{chart.chartType}</Tag>}
                        </Space>
                      }
                    >
                      <div className="chart-info">
                        <Paragraph ellipsis={{ rows: 2 }} type="secondary">
                          <strong>分析目标：</strong>
                          {chart.goal || '无'}
                        </Paragraph>
                      </div>

                      {chartOption && (
                        <div className="chart-container">
                          <ReactECharts option={chartOption} style={{ height: '300px' }} />
                        </div>
                      )}

                      {chart.genResult && (
                        <div className="chart-result">
                          <Paragraph
                            ellipsis={{ rows: 3, expandable: true, symbol: '展开' }}
                            style={{ marginTop: 16 }}
                          >
                            <strong>分析结论：</strong>
                            {chart.genResult}
                          </Paragraph>
                        </div>
                      )}
                    </Card>
                  </Col>
                );
              })}
            </Row>

            <div className="pagination-container">
              <Pagination
                current={searchParams.current}
                pageSize={searchParams.pageSize}
                total={total}
                showTotal={(total) => `共 ${total} 条`}
                showSizeChanger
                pageSizeOptions={['4', '8', '12', '16']}
                onChange={handlePageChange}
              />
            </div>
          </>
        )}
      </Spin>
    </div>
  );
};

export default MyChartPage;
