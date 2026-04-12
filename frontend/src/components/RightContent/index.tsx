import { QuestionCircleOutlined } from '@ant-design/icons';
import '@umijs/max';
export type SiderTheme = 'light' | 'dark';

// SelectLang 组件已移除，如需使用请安装 @umijs/max 的国际化插件
export const Question = () => {
  return (
    <div
      style={{
        display: 'flex',
        height: 26,
      }}
      onClick={() => {
        window.open('https://pro.ant.design/docs/getting-started');
      }}
    >
      <QuestionCircleOutlined />
    </div>
  );
};
