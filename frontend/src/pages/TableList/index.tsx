import {
  addChartUsingPost,
  deleteChartUsingPost,
  listChartByPageUsingPost,
} from '@/services/yubi/chartController';
import { PlusOutlined } from '@ant-design/icons';
import type { ActionType, ProColumns, ProDescriptionsItemProps } from '@ant-design/pro-components';
import {
  FooterToolbar,
  ModalForm,
  PageContainer,
  ProDescriptions,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import '@umijs/max';
import { Button, Drawer, message } from 'antd';
import React, { useRef, useState } from 'react';

const handleAdd = async (fields: API.ChartAddRequest) => {
  const hide = message.loading('正在添加');
  try {
    await addChartUsingPost({ ...fields });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error('添加失败，请重试');
    return false;
  }
};

// handleUpdate 函数已移除（未使用）

const handleRemove = async (selectedRows: API.Chart[]) => {
  const hide = message.loading('正在删除');
  if (!selectedRows) return true;
  try {
    for (const row of selectedRows) {
      await deleteChartUsingPost({ id: row.id });
    }
    hide();
    message.success('删除成功');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};

const TableList: React.FC = () => {
  const [createModalOpen, handleModalOpen] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<API.Chart>();
  const [selectedRowsState, setSelectedRows] = useState<API.Chart[]>([]);

  const columns: ProColumns<API.Chart>[] = [
    {
      title: '图表名称',
      dataIndex: 'name',
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: '分析目标',
      dataIndex: 'goal',
      valueType: 'textarea',
    },
    {
      title: '图表类型',
      dataIndex: 'chartType',
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      valueType: 'dateTime',
      sorter: true,
      hideInSearch: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => [
        <a
          key="detail"
          onClick={() => {
            setCurrentRow(record);
            setShowDetail(true);
          }}
        >
          查看
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<API.Chart>
        headerTitle={'图表列表'}
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              handleModalOpen(true);
            }}
          >
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={async (params, sort) => {
          const res = await listChartByPageUsingPost({
            ...params,
            sortField: sort && Object.keys(sort)[0],
            sortOrder: sort && Object.values(sort)[0],
          } as API.ChartQueryRequest);
          return {
            data: res?.data?.records ?? [],
            success: true,
            total: res?.data?.total ?? 0,
          };
        }}
        columns={columns}
        rowSelection={{
          onChange: (_, selectedRows) => {
            setSelectedRows(selectedRows);
          },
        }}
      />
      {selectedRowsState?.length > 0 && (
        <FooterToolbar
          extra={
            <div>
              已选择 <a style={{ fontWeight: 600 }}>{selectedRowsState.length}</a> 项
            </div>
          }
        >
          <Button
            onClick={async () => {
              await handleRemove(selectedRowsState);
              setSelectedRows([]);
              actionRef.current?.reloadAndRest?.();
            }}
          >
            批量删除
          </Button>
        </FooterToolbar>
      )}
      <ModalForm
        title={'新建图表'}
        width="400px"
        open={createModalOpen}
        onOpenChange={handleModalOpen}
        onFinish={async (value) => {
          const success = await handleAdd(value as API.ChartAddRequest);
          if (success) {
            handleModalOpen(false);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
      >
        <ProFormText
          rules={[{ required: true, message: '图表名称为必填项' }]}
          width="md"
          name="name"
          label="图表名称"
        />
        <ProFormTextArea width="md" name="goal" label="分析目标" />
      </ModalForm>

      <Drawer
        width={600}
        open={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.name && (
          <ProDescriptions<API.Chart>
            column={2}
            title={currentRow?.name}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<API.Chart>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};
export default TableList;
