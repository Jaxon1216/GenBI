import { Card, Col, Row, Skeleton, Space } from 'antd';
import React from 'react';

interface ChartCardSkeletonProps {
  count?: number;
}

const ChartCardSkeleton: React.FC<ChartCardSkeletonProps> = ({ count = 4 }) => (
  <Row gutter={[16, 16]}>
    {Array.from({ length: count }).map((_, i) => (
      <Col key={i} xs={24} sm={24} md={24} lg={12} xl={12} xxl={12}>
        <Card>
          <Space style={{ marginBottom: 16 }}>
            <Skeleton.Avatar active size="default" />
            <div>
              <Skeleton.Input active size="small" style={{ width: 120 }} />
              <br />
              <Skeleton.Input active size="small" style={{ width: 80, marginTop: 4 }} />
            </div>
          </Space>
          <Skeleton.Node active style={{ width: '100%', height: 180 }}>
            <div style={{ width: '100%' }} />
          </Skeleton.Node>
        </Card>
      </Col>
    ))}
  </Row>
);

export default ChartCardSkeleton;
