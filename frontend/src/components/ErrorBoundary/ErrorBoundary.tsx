import { tx } from '@/i18n/text';
import React from 'react';
import { Result, Button } from 'antd';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

/**
 * 错误边界 — 捕获子组件渲染错误，显示 fallback UI
 */
class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    // 生产环境可接入日志服务
    console.error('ErrorBoundary caught:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: undefined });
  };

  render(): React.ReactNode {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <Result
          status="error"
          title={tx('页面出错了')}
          subTitle={this.state.error?.message || tx('发生了未知错误')}
          extra={
            <Button type="primary" onClick={this.handleReset}>
              {tx('重试')}
            </Button>
          }
        />
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
