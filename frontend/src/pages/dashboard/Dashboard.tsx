import React from 'react';
import { Row, Col, Card, Statistic, List, Tag, Progress } from 'antd';
import {
  UserOutlined,
  TeamOutlined,
  ScheduleOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';

const Dashboard: React.FC = () => {
  // 统计卡片数据
  const stats = [
    { title: '会员总数', value: 328, icon: <UserOutlined style={{ color: '#1677ff' }} />, color: '#e6f4ff' },
    { title: '待面试', value: 15, icon: <ScheduleOutlined style={{ color: '#faad14' }} />, color: '#fffbe6' },
    { title: '进行中任务', value: 23, icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />, color: '#f6ffed' },
    { title: '本月会议', value: 8, icon: <TeamOutlined style={{ color: '#722ed1' }} />, color: '#f9f0ff' },
  ];

  // 会员增长趋势图
  const memberGrowthOption = {
    title: { text: '会员增长趋势', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月'],
    },
    yAxis: { type: 'value' },
    series: [
      {
        data: [50, 80, 120, 150, 180, 220, 280, 328],
        type: 'line',
        smooth: true,
        areaStyle: { opacity: 0.3 },
        lineStyle: { color: '#1677ff' },
        itemStyle: { color: '#1677ff' },
      },
    ],
    grid: { left: 40, right: 20, top: 40, bottom: 30 },
  };

  // 部门分布饼图
  const departmentOption = {
    title: { text: '部门人员分布', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'item' },
    legend: { bottom: 10, left: 'center' },
    series: [
      {
        type: 'pie',
        radius: ['40%', '65%'],
        avoidLabelOverlap: false,
        label: { show: false },
        data: [
          { value: 45, name: '技术部' },
          { value: 38, name: '宣传部' },
          { value: 32, name: '组织部' },
          { value: 28, name: '外联部' },
          { value: 25, name: '办公室' },
        ],
      },
    ],
  };

  // 最近活动
  const recentActivities = [
    { user: '张三', action: '提交了入会申请', time: '10分钟前', type: '申请' },
    { user: '李四', action: '完成了面试安排', time: '30分钟前', type: '面试' },
    { user: '王五', action: '发起了新会议', time: '1小时前', type: '会议' },
    { user: '赵六', action: '更新了任务状态', time: '2小时前', type: '任务' },
    { user: '孙七', action: '提交了实习报告', time: '3小时前', type: '实习' },
  ];

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        {stats.map((stat, index) => (
          <Col span={6} key={index}>
            <Card>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Statistic title={stat.title} value={stat.value} />
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: 8,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 24,
                    background: stat.color,
                  }}
                >
                  {stat.icon}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 图表区域 */}
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={16}>
          <Card title="数据概览">
            <ReactECharts option={memberGrowthOption} style={{ height: 300 }} />
          </Card>
        </Col>
        <Col span={8}>
          <Card title="部门分布">
            <ReactECharts option={departmentOption} style={{ height: 300 }} />
          </Card>
        </Col>
      </Row>

      {/* 最近活动 & 待办事项 */}
      <Row gutter={16}>
        <Col span={12}>
          <Card title="最近活动">
            <List
              dataSource={recentActivities}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={
                      <span>
                        <strong>{item.user}</strong> {item.action}
                      </span>
                    }
                    description={item.time}
                  />
                  <Tag color="blue">{item.type}</Tag>
                </List.Item>
              )}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="任务进度">
            <List
              dataSource={[
                { name: '招新系统开发', progress: 80 },
                { name: '面试流程优化', progress: 60 },
                { name: '官网改版', progress: 45 },
                { name: '培训资料整理', progress: 30 },
              ]}
              renderItem={(item) => (
                <List.Item>
                  <div style={{ width: '100%' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                      <span>{item.name}</span>
                      <span>{item.progress}%</span>
                    </div>
                    <Progress percent={item.progress} size="small" />
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
