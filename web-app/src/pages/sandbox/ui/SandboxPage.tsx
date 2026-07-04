import { Icon } from "@iconify/react";
import { Link } from "react-router";
import { AppHeader } from "@widgets/header";
import { useReportList } from "@entities/report";
import type { Report } from "@shared/api";
import styles from "./SandboxPage.module.css";

const reportStatusIcons = {
  pending: "tabler:loader-2",
  passed: "tabler:circle-check-filled",
  failed: "tabler:circle-x-filled",
} as const;

const reportStatusText: Record<Report["status"], string> = {
  pending: "Processing",
  passed: "Passed",
  failed: "Failed",
};

function ReportStatusIcon({ status }: { status: Report["status"] }) {
  return (
    <span className={`${styles.statusIcon} ${styles[status]}`} aria-hidden="true">
      <Icon icon={reportStatusIcons[status]} width={34} height={34} />
    </span>
  );
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffH = Math.floor(diffMin / 60);
  if (diffH < 24) return `${diffH}h ago`;
  const diffD = Math.floor(diffH / 24);
  return `${diffD}d ago`;
}

function extractTaskName(uid: string): string {
  const slash = uid.indexOf("/");
  if (slash < 0) return uid;
  const repo = uid.slice(slash + 1);
  return repo.startsWith("golden-") ? repo.slice(7) : repo;
}

function ReportCard({ report }: { report: Report }) {
  const taskName = extractTaskName(report.uid);
  const commit = report.commit?.slice(0, 7) ?? "—";
  const s = report.summary;

  return (
    <Link
      to={`/sandbox/tasks/reports/${report.id}`}
      className={styles.reportCard}
      aria-label={`Открыть отчёт: ${taskName}`}
    >
      <span className={styles.cardBody}>
        <span className={styles.cardHeader}>
          <span className={styles.taskName}>{taskName}</span>
          <span className={styles.statusWrap} aria-label={reportStatusText[report.status]}>
            <ReportStatusIcon status={report.status} />
          </span>
        </span>

        <span className={styles.cardMessage}>{s.message}</span>

        <span className={styles.cardMeta}>
          <span className={styles.scoreGroup}>
            <span className={styles.scorePassed}>{s.passed} passed</span>
            <span className={styles.scoreFailed}>{s.failed} failed</span>
            {s.blocked > 0 && <span className={styles.scoreBlocked}>{s.blocked} blocked</span>}
            {s.warnings > 0 && <span className={styles.scoreWarnings}>{s.warnings} warn</span>}
          </span>
          <span className={styles.commitHash}>{commit}</span>
          <span className={styles.commitTime}>{formatTime(report.created_at)}</span>
        </span>
      </span>
    </Link>
  );
}

export function SandboxPage() {
  const { data, isLoading, error, fetchNextPage, hasNextPage, isFetchingNextPage } = useReportList({
    limit: 20,
  });
  const reports = data?.pages.flatMap((p) => p.items) ?? [];

  return (
    <>
      <AppHeader />
      <div className={styles.page}>
        <aside className={styles.sidebar} aria-label="Навигация SandBox">
          <nav className={styles.nav}>
            <Link to="/sandbox/tasks" className={`${styles.navItem} ${styles.navItemActive}`}>
              Задачи
            </Link>
          </nav>
        </aside>

        <main className={styles.workspace}>
          <div className={styles.contentHeader}>
            <h1 className={styles.title}>Задачи</h1>
            <p className={styles.description}>Последние CI-отчёты вашего аккаунта.</p>
          </div>

          <section aria-label="CI-отчёты">
            {isLoading && <div className={styles.empty}>Загрузка отчётов…</div>}
            {error && <div className={styles.empty}>Не удалось загрузить отчёты.</div>}
            {!isLoading && reports.length === 0 && (
              <div className={styles.empty}>Нет CI-отчётов.</div>
            )}
            <div className={styles.reportList}>
              {reports.map((report) => (
                <ReportCard key={report.id} report={report} />
              ))}
            </div>
            {hasNextPage && (
              <button
                type="button"
                className={styles.loadMore}
                disabled={isFetchingNextPage}
                onClick={() => fetchNextPage()}
              >
                {isFetchingNextPage ? "Загрузка…" : "Загрузить ещё"}
              </button>
            )}
          </section>
        </main>
      </div>
    </>
  );
}
