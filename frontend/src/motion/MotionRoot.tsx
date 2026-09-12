import React from 'react';
import { MotionConfig } from 'motion/react';
import { useThemeLang } from '@/theme/ThemeLangContext';
import { motionEase } from './tokens';
import { resolveMotionMode } from './preference';

const MotionRoot: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { reduceMotion } = useThemeLang();
  return (
    <MotionConfig
      reducedMotion={resolveMotionMode(reduceMotion)}
      transition={{ duration: 0.22, ease: motionEase }}
    >
      {children}
    </MotionConfig>
  );
};

export default MotionRoot;
