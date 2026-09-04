import React from 'react';
import { Result, Button } from 'antd';
import { HomeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

/**
 * 404 页面
 *
 * 访问不存在的路由时展示，提供返回首页入口。
 */
const NotFound: React.FC = () => {
  const navigate = useNavigate();

  return (
    <Result
      status="404"
      title="404"
      subTitle="抱歉，您访问的页面不存在或已被移除。"
      extra={
        <Button
          type="primary"
          icon={<HomeOutlined />}
          onClick={() => navigate('/', { replace: true })}
        >
          返回首页
        </Button>
      }
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
      }}
    />
  );
};

export default NotFound;
