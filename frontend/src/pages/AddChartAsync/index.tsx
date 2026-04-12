import { genChartByAiAsyncUsingPost } from '@/services/yubi/chartController';
import { UploadOutlined } from '@ant-design/icons';
import { history } from '@umijs/max';
import { Button, Card, Col, Form, Input, message, Row, Select, Space, Upload } from 'antd';
import { useForm } from 'antd/es/form/Form';
import TextArea from 'antd/es/input/TextArea';
import React, { useState } from 'react';

const AddChartAsync: React.FC = () => {
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [form] = useForm();

  const onFinish = async (values: any) => {
    if (submitting) return;
    setSubmitting(true);

    const params = {
      ...values,
      file: undefined,
    };
    try {
      const res = await genChartByAiAsyncUsingPost(params, {}, values.file.file.originFileObj);
      if (!res?.data) {
        message.error('分析失败');
      } else {
        message.success('提交成功，请到「我的图表」页面查看生成进度');
        form.resetFields();
        setTimeout(() => history.push('/my_chart'), 1500);
      }
    } catch (e: any) {
      message.error('分析失败，' + e.message);
    }
    setSubmitting(false);
  };

  return (
    <div className="add-chart-async">
      <Row gutter={24}>
        <Col span={12}>
          <Card title="智能分析（异步）">
            <Form
              form={form}
              name="addChartAsync"
              labelAlign="left"
              labelCol={{ span: 4 }}
              wrapperCol={{ span: 16 }}
              onFinish={onFinish}
              initialValues={{}}
            >
              <Form.Item
                name="goal"
                label="分析目标"
                rules={[{ required: true, message: '请输入分析目标!' }]}
              >
                <TextArea placeholder="请输入你的分析需求，比如：分析网站用户的增长情况" />
              </Form.Item>

              <Form.Item name="name" label="图表名称">
                <Input placeholder="请输入图表名称" />
              </Form.Item>

              <Form.Item name="chartType" label="图表类型">
                <Select
                  options={[
                    { value: '折线图', label: '折线图' },
                    { value: '柱状图', label: '柱状图' },
                    { value: '堆叠图', label: '堆叠图' },
                    { value: '饼图', label: '饼图' },
                    { value: '雷达图', label: '雷达图' },
                  ]}
                />
              </Form.Item>

              <Form.Item name="file" label="原始数据">
                <Upload name="file" maxCount={1}>
                  <Button icon={<UploadOutlined />}>上传 CSV 文件</Button>
                </Upload>
              </Form.Item>

              <Form.Item wrapperCol={{ span: 16, offset: 4 }}>
                <Space>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={submitting}
                    disabled={submitting}
                  >
                    提交
                  </Button>
                  <Button htmlType="reset">重置</Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="使用说明">
            <p>异步模式将在后台生成图表，提交后可在「我的图表」页面查看进度。</p>
            <p>适用于数据量较大、生成时间较长的分析场景。</p>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default AddChartAsync;
