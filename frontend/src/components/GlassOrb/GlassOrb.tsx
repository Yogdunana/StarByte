import styles from './GlassOrb.module.css';

/** Decorative CSS only: no canvas, animation loop, or pointer listeners. */
export default function GlassOrb() {
  return <div className={styles.art} aria-hidden="true">
    <span className={styles.ring} /><span className={styles.ringTwo} />
    <span className={styles.orb} /><span className={styles.spark} />
    <span className={styles.caption}>BUILT<br />TOGETHER.</span>
  </div>;
}
