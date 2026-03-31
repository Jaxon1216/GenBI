import { Modal, Typography, Tag, Empty } from 'antd';
import React, { useMemo } from 'react';
import { FixedSizeList as VirtualList } from 'react-window';

const { Text } = Typography;

interface VirtualDataTableProps {
  visible: boolean;
  onClose: () => void;
  csvData: string | undefined | null;
  chartName?: string;
}

function parseCsv(csv: string): { headers: string[]; rows: string[][] } {
  const lines = csv.split('\n').filter((line) => line.trim() !== '');
  if (lines.length === 0) return { headers: [], rows: [] };
  const headers = lines[0].split(',').map((h) => h.trim());
  const rows = lines.slice(1).map((line) => line.split(',').map((cell) => cell.trim()));
  return { headers, rows };
}

const ROW_HEIGHT = 40;
const MAX_VISIBLE_ROWS = 15;

const VirtualDataTable: React.FC<VirtualDataTableProps> = ({ visible, onClose, csvData, chartName }) => {
  const { headers, rows } = useMemo(() => parseCsv(csvData ?? ''), [csvData]);

  const listHeight = Math.min(rows.length, MAX_VISIBLE_ROWS) * ROW_HEIGHT;

  const renderRow = ({ index, style }: { index: number; style: React.CSSProperties }) => {
    const row = rows[index];
    return (
      <div
        style={{
          ...style,
          display: 'flex',
          alignItems: 'center',
          borderBottom: '1px solid #f0f0f0',
          padding: '0 8px',
          backgroundColor: index % 2 === 0 ? '#fafafa' : '#fff',
        }}
      >
        <Text style={{ width: 50, flexShrink: 0 }} type="secondary">
          {index + 1}
        </Text>
        {row.map((cell, ci) => (
          <Text
            key={ci}
            ellipsis={{ tooltip: cell }}
            style={{ flex: 1, minWidth: 80, padding: '0 4px' }}
          >
            {cell}
          </Text>
        ))}
      </div>
    );
  };

  return (
    <Modal
      title={
        <span>
          原始数据 {chartName && <Tag color="blue">{chartName}</Tag>}
          <Tag>{rows.length} 行</Tag>
        </span>
      }
      open={visible}
      onCancel={onClose}
      footer={null}
      width={Math.max(600, headers.length * 120 + 100)}
    >
      {headers.length === 0 ? (
        <Empty description="无数据" />
      ) : (
        <>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              padding: '8px',
              backgroundColor: '#e6f4ff',
              borderRadius: '4px 4px 0 0',
              fontWeight: 600,
            }}
          >
            <Text style={{ width: 50, flexShrink: 0 }} type="secondary">
              #
            </Text>
            {headers.map((h, i) => (
              <Text key={i} style={{ flex: 1, minWidth: 80, padding: '0 4px' }}>
                {h}
              </Text>
            ))}
          </div>
          <VirtualList
            height={listHeight || ROW_HEIGHT}
            itemCount={rows.length}
            itemSize={ROW_HEIGHT}
            width="100%"
          >
            {renderRow}
          </VirtualList>
        </>
      )}
    </Modal>
  );
};

export default VirtualDataTable;
