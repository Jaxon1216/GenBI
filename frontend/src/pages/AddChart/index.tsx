import { genChartByAiUsingPost } from '@/services/yubi/chartController';
import SafeChart from '@/components/SafeChart';
import { UploadOutlined } from '@ant-design/icons';
import { Button, Card, Col, Divider, Form, Input, message, Row, Select, Space, Spin, Upload } from 'antd';
import TextArea from 'antd/es/input/TextArea';
import React, { useState } from 'react';

/**
 * 添加图表页面
 * @constructor
 */
// 把多余的状态删掉，页面名称改为AddChart
const AddChart: React.FC = () => {
  const [chart, setChart] = useState<API.BiResponse>();
  const [rawChartJson, setRawChartJson] = useState<string>();
  const [submitting, setSubmitting] = useState<boolean>(false);

  /**
   * 提交
   * @param values
   */
    const onFinish = async (values: any) => {
      // 如果已经是提交中的状态(还在加载)，直接返回，避免重复提交
    if (submitting) {
      return;
    }
    // 当开始提交，把submitting设置为true
    setSubmitting(true);
    setChart(undefined);
    setRawChartJson(undefined);

      // 对接后端，上传数据
      const params = {
        ...values,
        file: undefined,
      };
      try {
        // 需要取到上传的原始数据file→file→originFileObj(原始数据)
        const res = await genChartByAiUsingPost(params, {}, values.file.file.originFileObj);
        // 正常情况下，如果没有返回值就分析失败，有，就分析成功
        if (!res?.data) {
          message.error('分析失败');
        } else {
          message.success('分析成功');
          setChart(res.data);
          setRawChartJson(res.data.genChart);
        }  
      // 异常情况下，提示分析失败+具体失败原因
      } catch (e: any) {
        message.error('分析失败,' + e.message);
      }
      // 当结束提交，把submitting设置为false
      setSubmitting(false);
    };  

  return (
    // 把页面内容指定一个类名add-chart
    <div className="add-chart">
      {/* 变成两列 gutter列与列之间的间隔*/}
      <Row gutter={24}>
        {/* 表单放在第一列,卡片组件里 */}
        <Col span={12}>
          <Card title="智能分析">
            <Form
            // 表单名称改为addChart
            name="addChart"
            // label标签的文本对齐方式
            labelAlign="left" 
            // label标签布局，同<Col>组件，设置 span offset 值，如 {span: 3, offset: 12} 
            labelCol={{ span: 4 }}
            // 设置控件布局样式
            wrapperCol={{ span: 16 }}
            onFinish={onFinish}
            // 初始化数据啥都不填，为空
            initialValues={{  }}
            >
            {/* 前端表单的name，对应后端接口请求参数里的字段，
            此处name对应后端分析目标goal,label是左侧的提示文本，
            rules=....是必填项提示*/}
            <Form.Item name="goal" label="分析目标" rules={[{ required: true, message: '请输入分析目标!' }]}>
                {/* placeholder文本框内的提示语 */}
                <TextArea placeholder="请输入你的分析需求，比如：分析网站用户的增长情况"/>
            </Form.Item>

            {/* 还要输入图表名称 */}
            <Form.Item name="name" label="图表名称">
                <Input placeholder="请输入图表名称" />
            </Form.Item>

            {/* 图表类型是非必填，所以不做校验 */}
            <Form.Item
              name="chartType"
              label="图表类型"
              >
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

            {/* 文件上传 */}
            <Form.Item
                name="file"
                label="原始数据"
              >
                {/* action:当你把文件上传之后，它会把文件上传至哪个接口。
                    这里肯定是调用自己的后端，先不用这个;
                    maxCount={1} 限制文件上传数量为1 */}
                <Upload name="file" maxCount={1}>
                  <Button icon={<UploadOutlined />}>上传 CSV 文件</Button>
                </Upload>
              </Form.Item>
              {/* offset设置和label标签一样的宽度，这样就能保持对齐；
                  其他占用的列设置成16 */}
              <Form.Item wrapperCol={{ span: 16, offset: 4 }}>
                <Space>
                  {/* 加个loading：就是把submitting的状态加入进来，
                  加个disabled：如果正在提交，就让这个按钮禁用，不允许重复点击*/}
                  <Button type="primary" htmlType="submit" loading={submitting} disabled={submitting}>
                    提交
                  </Button>
                  <Button htmlType="reset">重置</Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
               </Col>
        {/* 分析结论和图表放在第二列 */}
        <Col span={12}>
        <Card title="分析结论">
            {/* 如果分析结论存在，就展示分析结论；
                不存在则显示'请先在左侧进行提交' */}
            {chart?.genResult ?? <div>请先在左侧进行提交</div>}
             {/* 提交中，还未返回结果，分析结论就显示加载中的组件 */}
            <Spin spinning={submitting}/>
        </Card>
        {/* 加一个间距 */}
        <Divider />
        <Card title="可视化图表">
            {rawChartJson ? <SafeChart rawChartJson={rawChartJson} /> : <div>请先在左侧进行提交</div>}
            <Spin spinning={submitting}/>
        </Card>
        </Col>
      </Row>
    </div>
  );
};
export default AddChart;
