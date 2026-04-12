import ChartErrorBoundary from '@/components/ChartErrorBoundary';
import { parseChartOption } from '@/utils/chartSchema';
import { Alert, Button } from 'antd';
import ReactECharts from 'echarts-for-react';
import React, { useMemo } from 'react';

interface SafeChartProps {
  rawChartJson: string | undefined | null;
  stripTitle?: boolean;
  style?: React.CSSProperties;
  onRetry?: () => void;
}

const SafeChart: React.FC<SafeChartProps> = ({
  rawChartJson,
  stripTitle = false,
  style,
  onRetry,
}) => {
  const result = useMemo(
    () => parseChartOption(rawChartJson, stripTitle),
    [rawChartJson, stripTitle],
  );

  if (!result.success) {
    if (result.fallbackOption) {
      return (
        <>
          <Alert
            message="图表数据异常"
            description={result.error}
            type="warning"
            showIcon
            style={{ marginBottom: 8 }}
            action={
              onRetry ? (
                <Button size="small" type="primary" onClick={onRetry}>
                  重试
                </Button>
              ) : undefined
            }
          />
          <ChartErrorBoundary onRetry={onRetry}>
            <ReactECharts option={result.fallbackOption} style={style} />
          </ChartErrorBoundary>
        </>
      );
    }
    return (
      <Alert
        message="图表渲染失败"
        description={result.error}
        type="error"
        showIcon
        action={
          onRetry ? (
            <Button size="small" type="primary" onClick={onRetry}>
              重试
            </Button>
          ) : undefined
        }
      />
    );
  }

  return (
    <ChartErrorBoundary onRetry={onRetry}>
      <ReactECharts option={result.option} style={style} />
    </ChartErrorBoundary>
  );
};

export default SafeChart;
