export default [
  { name:'登录',path: '/user', layout: false, routes: [{ path: '/user/login', component: './User/Login' }] },
  { name:'欢迎页面',path: '/welcome', icon: 'smile', component: './Welcome' },
  { name:'添加图表',path:'/add_chart',icon: 'signalFilled', component: './AddChart' },
  { name:'添加图表（异步版）',path:'/add_chart_async',icon: 'signalFilled', component: './AddChartAsync' },
  { name:'我的图表',path:'/my_chart',icon: 'appleOutlined', component: './myChart' },
  {
    path: '/admin',
    icon: 'crown',
    access: 'canAdmin',
    name: '管理页面',
    routes: [
      { path: '/admin', name: '管理页面', redirect: '/admin/sub-page' },
      { path: '/admin/sub-page', name: '管理页面2', component: './Admin' },
    ],
  },
  { name:'表格页面',icon: 'table', path: '/list', component: './TableList' },
  { path: '/', redirect: '/welcome' },
  { path: '*', layout: false, component: './404' },
];
