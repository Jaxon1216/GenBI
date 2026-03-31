import { useCallback, useEffect, useRef } from 'react';

interface UsePollingOptions {
  enabled: boolean;
  baseInterval?: number;
  maxInterval?: number;
  backoffFactor?: number;
}

/**
 * 智能轮询 Hook，支持指数退避和页面可见性感知。
 * - enabled=true 时开始轮询，false 时停止
 * - 页面不可见时暂停，恢复可见后立即执行一次
 * - 连续无变化时增大间隔（指数退避），有变化时重置为 baseInterval
 */
export function usePolling(callback: () => Promise<void>, options: UsePollingOptions) {
  const { enabled, baseInterval = 3000, maxInterval = 30000, backoffFactor = 1.5 } = options;

  const timerRef = useRef<ReturnType<typeof setTimeout>>();
  const currentIntervalRef = useRef(baseInterval);
  const previousDataRef = useRef<string>('');
  const isMountedRef = useRef(true);

  const clearTimer = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = undefined;
    }
  }, []);

  const scheduleNext = useCallback(() => {
    clearTimer();
    if (!isMountedRef.current || !enabled) return;
    timerRef.current = setTimeout(async () => {
      if (!isMountedRef.current || !enabled) return;
      await callback();
      scheduleNext();
    }, currentIntervalRef.current);
  }, [callback, enabled, clearTimer]);

  const reportDataSnapshot = useCallback(
    (snapshot: string) => {
      if (snapshot !== previousDataRef.current) {
        currentIntervalRef.current = baseInterval;
        previousDataRef.current = snapshot;
      } else {
        currentIntervalRef.current = Math.min(
          currentIntervalRef.current * backoffFactor,
          maxInterval,
        );
      }
    },
    [baseInterval, maxInterval, backoffFactor],
  );

  useEffect(() => {
    if (!enabled) {
      clearTimer();
      currentIntervalRef.current = baseInterval;
      return;
    }
    scheduleNext();
    return clearTimer;
  }, [enabled, scheduleNext, clearTimer, baseInterval]);

  useEffect(() => {
    const handleVisibility = () => {
      if (document.visibilityState === 'visible' && enabled && isMountedRef.current) {
        currentIntervalRef.current = baseInterval;
        callback().then(scheduleNext);
      } else {
        clearTimer();
      }
    };
    document.addEventListener('visibilitychange', handleVisibility);
    return () => document.removeEventListener('visibilitychange', handleVisibility);
  }, [enabled, callback, scheduleNext, clearTimer, baseInterval]);

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      clearTimer();
    };
  }, [clearTimer]);

  return { reportDataSnapshot };
}
