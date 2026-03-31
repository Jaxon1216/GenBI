/**
 * 生成测试图表数据并插入 MySQL。
 *
 * 用法：node scripts/generate_test_data.js [userId]
 * 默认 userId=1，可通过参数指定。
 *
 * 前置：npm install mysql2（一次性）
 */

const mysql = require('mysql2/promise');

const USER_ID = process.argv[2] || '1';
const CHART_COUNT = 20;

const CHART_TYPES = ['折线图', '柱状图', '饼图', '堆叠图', '雷达图'];

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function generateCsv(rows) {
  const months = [];
  for (let i = 1; i <= rows; i++) {
    const y = 2020 + Math.floor((i - 1) / 12);
    const m = ((i - 1) % 12) + 1;
    months.push(`${y}-${String(m).padStart(2, '0')}`);
  }
  const headers = '日期,用户数,活跃用户,新增用户,付费用户,收入(万元)';
  const lines = months.map((month) => {
    const users = randomInt(1000, 50000);
    const active = randomInt(Math.floor(users * 0.3), Math.floor(users * 0.8));
    const newUsers = randomInt(50, 3000);
    const paying = randomInt(10, Math.floor(users * 0.1));
    const revenue = (paying * randomInt(10, 200) / 100).toFixed(2);
    return `${month},${users},${active},${newUsers},${paying},${revenue}`;
  });
  return [headers, ...lines].join('\n');
}

function generateLineChart(name, csvRows) {
  const lines = csvRows.split('\n');
  const headers = lines[0].split(',');
  const xData = [];
  const series1 = [];
  const series2 = [];
  const limit = Math.min(lines.length, 50);
  for (let i = 1; i < limit; i++) {
    const cells = lines[i].split(',');
    xData.push(cells[0]);
    series1.push(parseInt(cells[1]));
    series2.push(parseInt(cells[2]));
  }
  return JSON.stringify({
    title: { text: name },
    tooltip: { trigger: 'axis' },
    legend: { data: [headers[1], headers[2]] },
    xAxis: { type: 'category', data: xData },
    yAxis: { type: 'value' },
    series: [
      { name: headers[1], type: 'line', data: series1, smooth: true },
      { name: headers[2], type: 'line', data: series2, smooth: true },
    ],
  });
}

function generateBarChart(name, csvRows) {
  const lines = csvRows.split('\n');
  const headers = lines[0].split(',');
  const xData = [];
  const data1 = [];
  const data2 = [];
  const limit = Math.min(lines.length, 25);
  for (let i = 1; i < limit; i++) {
    const cells = lines[i].split(',');
    xData.push(cells[0]);
    data1.push(parseInt(cells[3]));
    data2.push(parseInt(cells[4]));
  }
  return JSON.stringify({
    title: { text: name },
    tooltip: { trigger: 'axis' },
    legend: { data: [headers[3], headers[4]] },
    xAxis: { type: 'category', data: xData },
    yAxis: { type: 'value' },
    series: [
      { name: headers[3], type: 'bar', data: data1 },
      { name: headers[4], type: 'bar', data: data2 },
    ],
  });
}

function generatePieChart(name, csvRows) {
  const lines = csvRows.split('\n');
  const limit = Math.min(lines.length, 8);
  const pieData = [];
  for (let i = 1; i < limit; i++) {
    const cells = lines[i].split(',');
    pieData.push({ name: cells[0], value: parseInt(cells[1]) });
  }
  return JSON.stringify({
    title: { text: name },
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{ type: 'pie', radius: '60%', data: pieData }],
  });
}

function generateStackChart(name, csvRows) {
  const lines = csvRows.split('\n');
  const headers = lines[0].split(',');
  const xData = [];
  const s1 = [], s2 = [], s3 = [];
  const limit = Math.min(lines.length, 25);
  for (let i = 1; i < limit; i++) {
    const cells = lines[i].split(',');
    xData.push(cells[0]);
    s1.push(parseInt(cells[2]));
    s2.push(parseInt(cells[3]));
    s3.push(parseInt(cells[4]));
  }
  return JSON.stringify({
    title: { text: name },
    tooltip: { trigger: 'axis' },
    legend: { data: [headers[2], headers[3], headers[4]] },
    xAxis: { type: 'category', data: xData },
    yAxis: { type: 'value' },
    series: [
      { name: headers[2], type: 'bar', stack: 'total', data: s1 },
      { name: headers[3], type: 'bar', stack: 'total', data: s2 },
      { name: headers[4], type: 'bar', stack: 'total', data: s3 },
    ],
  });
}

function generateRadarChart(name) {
  const indicators = ['技术', '产品', '设计', '运营', '市场', '客服'];
  return JSON.stringify({
    title: { text: name },
    tooltip: {},
    radar: { indicator: indicators.map((i) => ({ name: i, max: 100 })) },
    series: [{
      type: 'radar',
      data: [
        { name: 'Q1', value: indicators.map(() => randomInt(30, 95)) },
        { name: 'Q2', value: indicators.map(() => randomInt(30, 95)) },
      ],
    }],
  });
}

const CHART_NAMES = [
  '网站用户增长趋势分析', '月度活跃用户统计', '各月新增与付费用户对比',
  '用户构成分布', '团队能力雷达图', '季度收入趋势',
  '用户留存率分析', '核心指标月度对比', '2024年业务概览',
  '付费转化率变化', '全年用户数据总览', '运营数据月报',
  '产品各维度评分', '用户分层分析', '年度关键指标看板',
  '获客渠道效果分析', '日活月活趋势', '用户生命周期价值',
  '收入结构分析', '流失用户画像',
];

const GOALS = [
  '分析网站用户的增长情况', '对比各月的活跃用户数', '分析新增用户与付费用户的关系',
  '了解用户的构成分布', '评估团队各维度的能力', '分析季度收入变化趋势',
  '分析用户留存情况', '对比核心业务指标', '总览2024年业务数据',
  '分析付费转化率的变化趋势', '查看全年用户数据', '生成运营月报图表',
  '评估产品各维度得分', '对用户进行分层分析', '生成年度指标看板',
  '分析不同渠道的获客效果', '对比日活和月活的趋势', '计算用户生命周期价值',
  '分析收入结构占比', '分析流失用户特征',
];

const CONCLUSIONS = [
  '从数据来看，用户数整体呈上升趋势，尤其在Q3增长最为显著，建议加大Q3的营销投入。',
  '活跃用户占比稳定在40%-60%之间，说明产品粘性较好，但仍有提升空间。',
  '新增用户与付费用户呈正相关，付费转化率约为3%-5%，建议优化新用户引导流程。',
  '用户主要集中在一线城市，占比超过60%，可以考虑拓展二三线城市市场。',
  '团队在技术和产品维度表现突出，但市场和运营能力相对薄弱，建议加强相关培训。',
  '收入在年末有明显增长，与促销活动密切相关，建议制定更精细的促销策略。',
];

async function main() {
  const conn = await mysql.createConnection({
    host: 'localhost',
    port: 3306,
    user: 'root',
    password: '',
    database: 'yubi',
  });

  console.log('Connected to MySQL');

  const generators = { '折线图': generateLineChart, '柱状图': generateBarChart, '饼图': generatePieChart, '堆叠图': generateStackChart, '雷达图': generateRadarChart };

  const inserts = [];
  for (let i = 0; i < CHART_COUNT; i++) {
    const chartType = CHART_TYPES[i % CHART_TYPES.length];
    const name = CHART_NAMES[i % CHART_NAMES.length];
    const goal = GOALS[i % GOALS.length];
    const conclusion = CONCLUSIONS[i % CONCLUSIONS.length];

    // 虚拟列表测试：部分图表的 CSV 数据特别多（500-5000行）
    const csvRowCount = i < 5 ? randomInt(2000, 5000) : randomInt(12, 120);
    const csvData = generateCsv(csvRowCount);
    const genChart = generators[chartType](name, csvData);

    // 大部分 succeed，留几条 wait/running/failed 测试轮询
    let status = 'succeed';
    let execMessage = null;
    if (i === CHART_COUNT - 1) { status = 'wait'; execMessage = '排队中，请耐心等候'; }
    else if (i === CHART_COUNT - 2) { status = 'running'; execMessage = '正在生成图表...'; }
    else if (i === CHART_COUNT - 3) { status = 'failed'; execMessage = 'AI 生成超时，请重试'; }

    inserts.push([goal, name, csvData, chartType, genChart, conclusion, status, execMessage, USER_ID]);
  }

  const sql = `INSERT INTO chart (goal, name, chartData, chartType, genChart, genResult, status, execMessage, userId)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`;

  for (const values of inserts) {
    await conn.execute(sql, values);
  }

  console.log(`✅ 成功插入 ${CHART_COUNT} 条测试图表数据（userId=${USER_ID}）`);
  console.log(`   - 前 5 条包含 2000-5000 行 CSV 数据（用于测试虚拟列表）`);
  console.log(`   - 最后 3 条分别是 failed / running / wait 状态（用于测试轮询）`);
  console.log(`   - 其余为 succeed 状态（用于测试看板拖拽）`);

  await conn.end();
}

main().catch((err) => {
  console.error('❌ 出错:', err.message);
  process.exit(1);
});
