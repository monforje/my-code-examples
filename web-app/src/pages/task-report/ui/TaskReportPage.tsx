import { Link, useNavigate, useParams } from "react-router";
import { Icon } from "@iconify/react";
import { useAuth } from "@features/auth";
import { useReportDetail, type ReportStep } from "@entities/report";
import { AppHeader } from "@widgets/header";
import { Alert, Spinner } from "@shared/ui";
import styles from "./TaskReportPage.module.css";

const statusMeta = {
  pending: {
    label: "В процессе",
    tone: "pending",
    icon: "tabler:loader-2",
    description: "CI ещё выполняется или ожидает результата.",
  },
  passed: {
    label: "Пройдено",
    tone: "passed",
    icon: "tabler:circle-check-filled",
    description: "Проверки прошли, критичных ошибок не найдено.",
  },
  failed: {
    label: "Провалено",
    tone: "failed",
    icon: "tabler:circle-x-filled",
    description: "CI нашёл ошибки, которые нужно исправить.",
  },
} as const;

const stepMeta = {
  passed: { label: "Пройден", tone: "passed", icon: "tabler:check" },
  failed: { label: "Ошибка", tone: "failed", icon: "tabler:x" },
  blocked: { label: "Блок", tone: "blocked", icon: "tabler:lock" },
  warning: { label: "Warning", tone: "warning", icon: "tabler:alert-triangle" },
} as const;

const dateFormatter = new Intl.DateTimeFormat("ru-RU", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return dateFormatter.format(date);
}

function shortCommit(commit: string) {
  return commit.length > 10 ? commit.slice(0, 10) : commit;
}

function extractTaskName(uid: string): string {
  const slash = uid.indexOf("/");
  if (slash < 0) return uid;
  const repo = uid.slice(slash + 1);
  return repo.startsWith("golden-") ? repo.slice(7) : repo;
}

function getStepMeta(status: ReportStep["status"]) {
  return stepMeta[status] ?? stepMeta.blocked;
}

export function TaskReportPage() {
  const { reportId } = useParams<{ reportId: string }>();
  const { user } = useAuth();
  const navigate = useNavigate();
  const { data: report, isLoading, isError } = useReportDetail(
    reportId || "",
    !!user && !!reportId,
  );

  const meta = report ? statusMeta[report.status] : statusMeta.pending;
  const taskName = report ? extractTaskName(report.uid) : "CI отчёт";

  return (
    <>
      <AppHeader />
      <main className={styles.page}>
        <div className={styles.gridBg} aria-hidden="true" />
        <div className={`${styles.container} cd-container`}>
          <button type="button" className={styles.backButton} onClick={() => navigate(-1)}>
            <Icon icon="tabler:arrow-left" width={18} />
            Назад
          </button>

          {isLoading && (
            <div className={styles.loadingCard}>
              <Spinner size={22} />
              <span>Загружаем CI-отчёт</span>
            </div>
          )}

          {isError && <Alert variant="danger">Не удалось загрузить CI-отчёт.</Alert>}

          {!isLoading && !isError && !report && (
            <section className={styles.emptyState}>
              <Icon icon="tabler:report-search" width={42} />
              <h1>Отчёт не найден</h1>
              <p>В текущей выдаче API нет отчёта с таким идентификатором.</p>
              <Link className="cd-button cd-button-soft" to="/sandbox/tasks">
                К списку отчётов
              </Link>
            </section>
          )}

          {report && (
            <>
              <section className={`${styles.hero} ${styles[meta.tone]}`}>
                <div className={styles.heroCopy}>
                  <div className={styles.kicker}>CI отчёт</div>
                  <h1>{taskName}</h1>
                  <p>{report.summary.message}</p>
                  <div className={styles.heroActions}>
                    <Link className="cd-button cd-button-soft" to="/sandbox/tasks">
                      Все отчёты
                    </Link>
                  </div>
                </div>
                <aside className={styles.statusPanel} aria-label="Статус отчёта">
                  <span className={styles.statusIcon}>
                    <Icon icon={meta.icon} width={34} />
                  </span>
                  <div>
                    <div className={styles.statusLabel}>{meta.label}</div>
                    <p>{meta.description}</p>
                  </div>
                  <div className={styles.reportMeta}>
                    <span>
                      <Icon icon="tabler:git-commit" width={16} />
                      {shortCommit(report.commit)}
                    </span>
                    <span>
                      <Icon icon="tabler:clock" width={16} />
                      {formatDate(report.created_at)}
                    </span>
                    <span>
                      <Icon icon="tabler:id" width={16} />
                      {report.id}
                    </span>
                  </div>
                </aside>
              </section>

              <div className={styles.reportLayoutSingle}>
                <section className={styles.mainColumn}>
                  <div className={styles.statsGrid}>
                    <Metric label="Passed" value={report.summary.passed} tone="passed" />
                    <Metric label="Failed" value={report.summary.failed} tone="failed" />
                    <Metric label="Blocked" value={report.summary.blocked} tone="blocked" />
                    <Metric label="Warnings" value={report.summary.warnings} tone="warning" />
                  </div>

                  {report.summary.root_cause && (
                    <section className={`${styles.card} ${styles.rootCause}`}>
                      <div className={styles.cardHeader}>
                        <Icon icon="tabler:bug" width={20} />
                        <h2>Root cause</h2>
                      </div>
                      <p>{report.summary.root_cause}</p>
                    </section>
                  )}

                  <section className={styles.card}>
                    <div className={styles.cardHeader}>
                      <Icon icon="tabler:timeline-event" width={20} />
                      <h2>Шаги проверки</h2>
                    </div>
                    <div className={styles.stepsList}>
                      {report.steps.map((step) => (
                        <StepCard key={`${step.index}-${step.name}`} step={step} />
                      ))}
                    </div>
                  </section>

                  {report.lint_errors.length > 0 && (
                    <section className={styles.card}>
                      <div className={styles.cardHeader}>
                        <Icon icon="tabler:code-dots" width={20} />
                        <h2>Lint errors</h2>
                      </div>
                      <div className={styles.lintList}>
                        {report.lint_errors.map((error, index) => (
                          <article key={`${error.file}-${error.line}-${error.col}-${index}`}>
                            <div className={styles.lintPath}>
                              {error.file}:{error.line}:{error.col}
                            </div>
                            <p>{error.message}</p>
                            <span>{error.rule}</span>
                          </article>
                        ))}
                      </div>
                    </section>
                  )}
                </section>

                <aside className={styles.sideColumn}>
                  <section className={styles.card}>
                    <div className={styles.cardHeader}>
                      <Icon icon="tabler:alert-triangle" width={20} />
                      <h2>Warnings</h2>
                    </div>
                    {report.warnings.length > 0 ? (
                      <ul className={styles.warningList}>
                        {report.warnings.map((warning, index) => (
                          <li key={`${warning}-${index}`}>{warning}</li>
                        ))}
                      </ul>
                    ) : (
                      <p className={styles.emptyText}>Предупреждений нет.</p>
                    )}
                  </section>

                  <section className={styles.card}>
                    <div className={styles.cardHeader}>
                      <Icon icon="tabler:file-analytics" width={20} />
                      <h2>Детали</h2>
                    </div>
                    <dl className={styles.detailsList}>
                      <div>
                        <dt>UID</dt>
                        <dd>{report.uid}</dd>
                      </div>
                      <div>
                        <dt>Raw log</dt>
                        <dd>{report.raw_log_available ? "Доступен" : "Недоступен"}</dd>
                      </div>
                    </dl>
                  </section>
                </aside>
              </div>
            </>
          )}
        </div>
      </main>
    </>
  );
}

function Metric({ label, value, tone }: { label: string; value: number; tone: keyof typeof styles }) {
  return (
    <article className={`${styles.metric} ${styles[tone]}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  );
}

function StepCard({ step }: { step: ReportStep }) {
  const meta = getStepMeta(step.status);

  return (
    <article className={`${styles.stepCard} ${styles[meta.tone]}`}>
      <div className={styles.stepTopline}>
        <span className={styles.stepIndex}>{step.index}</span>
        <h3>{step.name}</h3>
        <span className={styles.stepStatus}>
          <Icon icon={meta.icon} width={15} />
          {meta.label}
        </span>
      </div>
      {(step.http_status || step.code) && (
        <div className={styles.stepMeta}>
          {step.http_status && <span>HTTP {step.http_status}</span>}
          {step.code && <code>{step.code}</code>}
        </div>
      )}
      {step.failure && <p>{step.failure}</p>}
    </article>
  );
}
