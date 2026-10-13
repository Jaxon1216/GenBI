import { useEffect, useRef } from 'react';

/**
 * SSE 实时推送 Hook：为每个处于 pending（wait/running）的图表建立一条 EventSource 连接，
 * 收到后端经 Redis Pub/Sub 下发的状态变更事件后回调上层刷新；收到 succeed/failed 即关闭该连接。
 *
 * 降级策略：当浏览器不支持 EventSource，或所有连接均在超时窗口内未建立/断连时，
 * 通过 onFallback(true) 通知上层回退到 usePolling；SSE 恢复可用时回调 onFallback(false)。
 */

interface UseChartEventsOptions {
  // 是否启用 SSE（通常与「存在 pending 图表」联动）
  enabled: boolean;
  // SSE 不可用/断连时的降级信号：true 表示应启用轮询兜底
  onFallback?: (shouldFallback: boolean) => void;
  // 单条连接建立超时（毫秒）：超时未 open 视为不可用，触发降级
  connectTimeout?: number;
}

type ChartEventPayload = { chartId: number; status: string };

export function useChartEvents(
  chartIds: number[],
  onEvent: (chartId: number, status: string) => void,
  options: UseChartEventsOptions,
) {
  const { enabled, onFallback, connectTimeout = 10000 } = options;

  // 以 chartId 为键维护活跃连接，避免同一图表重复建连
  const sourcesRef = useRef<Map<number, EventSource>>(new Map());
  // 每条连接的建连超时定时器
  const timersRef = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map());
  // 用 ref 持有最新回调，避免因回调引用变化反复重建连接
  const onEventRef = useRef(onEvent);
  const onFallbackRef = useRef(onFallback);
  onEventRef.current = onEvent;
  onFallbackRef.current = onFallback;

  // 用 join 后的字符串作为依赖，内容不变则不重建连接
  const idsKey = chartIds.join(',');

  useEffect(() => {
    const timers = timersRef.current;
    const sources = sourcesRef.current;

    const closeSource = (id: number) => {
      const es = sources.get(id);
      if (es) {
        es.close();
        sources.delete(id);
      }
      const timer = timers.get(id);
      if (timer) {
        clearTimeout(timer);
        timers.delete(id);
      }
    };

    const closeAll = () => {
      sources.forEach((es) => es.close());
      sources.clear();
      timers.forEach((timer) => clearTimeout(timer));
      timers.clear();
    };

    // 未启用或浏览器不支持 EventSource：清理连接并要求降级轮询
    if (!enabled) {
      closeAll();
      return;
    }
    if (typeof EventSource === 'undefined') {
      closeAll();
      onFallbackRef.current?.(true);
      return;
    }

    const targetIds = idsKey
      ? idsKey
          .split(',')
          .map(Number)
          .filter((n) => n > 0)
      : [];

    // 关闭已不在 pending 列表中的连接
    sources.forEach((_es, id) => {
      if (!targetIds.includes(id)) closeSource(id);
    });

    // 为新的 pending 图表建立连接
    targetIds.forEach((id) => {
      if (sources.has(id)) return;

      // 同源请求，浏览器自动带上 cookie 完成鉴权
      const es = new EventSource(`/api/chart/sse?id=${id}`);
      sources.set(id, es);

      // 建连超时：超时仍未 open，视为 SSE 不可用，触发降级
      const timer = setTimeout(() => {
        if (sources.get(id) === es && es.readyState !== EventSource.OPEN) {
          closeSource(id);
          onFallbackRef.current?.(true);
        }
      }, connectTimeout);
      timers.set(id, timer);

      es.onopen = () => {
        // 连接成功：取消超时定时器，并解除降级（SSE 可用）
        const t = timers.get(id);
        if (t) {
          clearTimeout(t);
          timers.delete(id);
        }
        onFallbackRef.current?.(false);
      };

      es.onmessage = (ev) => {
        try {
          const payload: ChartEventPayload = JSON.parse(ev.data);
          onEventRef.current(payload.chartId, payload.status);
          // 终态后后端会结束流，这里主动关闭连接释放资源
          if (payload.status === 'succeed' || payload.status === 'failed') {
            closeSource(id);
          }
        } catch {
          // 坏消息忽略，不影响连接
        }
      };

      es.onerror = () => {
        // 断连/出错：关闭该连接并触发降级，由 usePolling 兜底
        closeSource(id);
        onFallbackRef.current?.(true);
      };
    });

    return closeAll;
  }, [enabled, idsKey, connectTimeout]);
}
