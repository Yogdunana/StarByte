import React from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import styles from './AboutLanding.module.css';

const CENTERS = [
  {
    key: 'rd',
    depts: ['innovation', 'project', 'curriculum'] as const,
  },
  {
    key: 'ops',
    depts: ['admin', 'brand'] as const,
  },
  {
    key: 'ext',
    depts: ['strategy', 'events'] as const,
  },
] as const;

const ROLE_GROUPS = [
  {
    label: 'roleGroupAssociation',
    roles: [
      'member',
      'probationOfficer',
      'officer',
      'viceMinister',
      'minister',
      'viceDirector',
      'director',
      'vicePresident',
      'president',
      'advisor',
      'honorary',
    ] as const,
  },
  {
    label: 'roleGroupTeam',
    roles: ['captain', 'teammate'] as const,
  },
  {
    label: 'roleGroupSystem',
    roles: ['admin'] as const,
  },
] as const;

const AboutLanding: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className={styles.page}>
      <section className={styles.hero}>
        <div className={styles.heroCopy}>
          <span className={styles.kicker}>{t('aboutUs.kicker')}</span>
          <h1>{t('aboutUs.title')}</h1>
          <div className={styles.names}>
            <span>SMBU-CA / SMBUCA</span>
            <span>КА МГУ-ППИ</span>
          </div>
          <p className={styles.lead}>{t('aboutUs.lead')}</p>
          <div className={styles.actions}>
            <Link className={styles.primary} to="/docs/association-charter">
              {t('aboutUs.readCharter')}
            </Link>
            <Link className={styles.ghost} to="/member/application">
              {t('aboutUs.apply')}
            </Link>
          </div>
        </div>
        <aside className={styles.heroAside}>
          <h2>{t('aboutUs.heroAsideTitle')}</h2>
          <p>{t('aboutUs.heroAsideLead')}</p>
          <ol>
            <li>{t('aboutUs.center.rd')}</li>
            <li>{t('aboutUs.center.ops')}</li>
            <li>{t('aboutUs.center.ext')}</li>
          </ol>
          <p>{t('aboutUs.heroAsideTeam')}</p>
        </aside>
      </section>

      <div className={styles.stats}>
        {(['centers', 'departments', 'teams', 'task'] as const).map((key) => (
          <div className={styles.stat} key={key}>
            <strong>{t(`aboutUs.stat.${key}Value`)}</strong>
            <span>{t(`aboutUs.stat.${key}`)}</span>
          </div>
        ))}
      </div>

      <section className={styles.section}>
        <h2>{t('aboutUs.orgTitle')}</h2>
        <p>{t('aboutUs.orgLead')}</p>
        <div className={styles.centers}>
          {CENTERS.map((center) => (
            <article className={styles.card} key={center.key}>
              <h3>{t(`aboutUs.center.${center.key}`)}</h3>
              <ul>
                {center.depts.map((dept) => (
                  <li key={dept}>{t(`aboutUs.dept.${dept}`)}</li>
                ))}
              </ul>
            </article>
          ))}
        </div>
      </section>

      <section className={styles.section}>
        <h2>{t('aboutUs.joinTitle')}</h2>
        <p>{t('aboutUs.joinLead')}</p>
        <div className={styles.tracks}>
          <article className={styles.track}>
            <h3>{t('aboutUs.memberTrackTitle')}</h3>
            <p>{t('aboutUs.memberTrackBody')}</p>
          </article>
          <article className={styles.track}>
            <h3>{t('aboutUs.officerTrackTitle')}</h3>
            <ol>
              <li>{t('aboutUs.officerStep1')}</li>
              <li>{t('aboutUs.officerStep2')}</li>
              <li>{t('aboutUs.officerStep3')}</li>
              <li>{t('aboutUs.officerStep4')}</li>
            </ol>
          </article>
        </div>
      </section>

      <section className={styles.section}>
        <h2>{t('aboutUs.teamsTitle')}</h2>
        <p>{t('aboutUs.teamsLead')}</p>
        <div className={styles.tracks}>
          <article className={styles.track}>
            <h3>{t('aboutUs.teamCaptain')}</h3>
            <p>{t('aboutUs.teamCaptainBody')}</p>
          </article>
          <article className={styles.track}>
            <h3>{t('aboutUs.teamMember')}</h3>
            <p>{t('aboutUs.teamMemberBody')}</p>
          </article>
        </div>
      </section>

      <section className={styles.section}>
        <h2>{t('aboutUs.rolesTitle')}</h2>
        <p>{t('aboutUs.rolesLead')}</p>
        <div className={styles.roleGroups}>
          {ROLE_GROUPS.map((group) => (
            <div className={styles.roleGroup} key={group.label}>
              <span className={styles.roleGroupLabel}>{t(`aboutUs.${group.label}`)}</span>
              <div className={styles.roles}>
                {group.roles.map((role) => (
                  <span className={styles.role} key={role}>
                    {t(`aboutUs.role.${role}`)}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </section>

      <p className={styles.note}>{t('aboutUs.note')}</p>
    </div>
  );
};

export default AboutLanding;
