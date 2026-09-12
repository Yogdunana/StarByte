import type { Transition, Variants } from 'motion/react';

/** Shared easing — matches --sb-ease. Only transform/opacity. */
export const motionEase: [number, number, number, number] = [0.2, 0.7, 0.3, 1];

export const shellTransition: Transition = {
  duration: 0.22,
  ease: motionEase,
};

export const enterTransition: Transition = {
  duration: 0.32,
  ease: motionEase,
};

export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: enterTransition },
};

export const fadeRight: Variants = {
  hidden: { opacity: 0, x: -8 },
  show: { opacity: 1, x: 0, transition: shellTransition },
};

export const staggerEnter: Variants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: 0.055,
      delayChildren: 0.04,
    },
  },
};

export const instantTransition: Transition = { duration: 0 };
