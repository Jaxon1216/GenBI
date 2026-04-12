import { Card, Col, Row, Skeleton } from 'antd';
import React from 'react';

const PageLoading: React.FC = () => (
  <div style={{ padding: 24 }}>
    <Row gutter={24}>
      <Col span={12}>
        <Card>
          <Skeleton active paragraph={{ rows: 6 }} />
        </Card>
      </Col>
      <Col span={12}>
        <Card>
          <Skeleton active paragraph={{ rows: 4 }} />
        </Card>
        <div style={{ marginTop: 16 }} />
        <Card>
          <Skeleton.Node active style={{ width: '100%', height: 200 }}>
            <div style={{ width: '100%' }} />
          </Skeleton.Node>
        </Card>
      </Col>
    </Row>
  </div>
);

export default PageLoading;
