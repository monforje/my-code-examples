import { lazy, Suspense } from "react";
import { Routes, Route } from "react-router";
import { AuthProvider } from "@features/auth";
import { ProtectedRoute } from "@shared/lib/protected-route";
import { ToastContainer } from "@shared/ui/Toast";
import { ToastProvider, useToast } from "@shared/lib/use-toast";

const TasksListPage = lazy(() =>
  import("@pages/tasks").then((m) => ({ default: m.TasksListPage })),
);
const TaskDetailPage = lazy(() =>
  import("@pages/task-detail").then((m) => ({ default: m.TaskDetailPage })),
);
const TaskReportPage = lazy(() =>
  import("@pages/task-report").then((m) => ({ default: m.TaskReportPage })),
);
const SandboxPage = lazy(() => import("@pages/sandbox").then((m) => ({ default: m.SandboxPage })));
const LoginPage = lazy(() => import("@pages/login").then((m) => ({ default: m.LoginPage })));
const RegisterPage = lazy(() =>
  import("@pages/register").then((m) => ({ default: m.RegisterPage })),
);
const VerifyPage = lazy(() => import("@pages/verify").then((m) => ({ default: m.VerifyPage })));
const ForgotPage = lazy(() => import("@pages/forgot").then((m) => ({ default: m.ForgotPage })));
const ForgotVerifyPage = lazy(() =>
  import("@pages/forgot-verify").then((m) => ({ default: m.ForgotVerifyPage })),
);
const ResetPage = lazy(() => import("@pages/reset").then((m) => ({ default: m.ResetPage })));
const CliLoginPage = lazy(() =>
  import("@pages/cli-login").then((m) => ({ default: m.CliLoginPage })),
);
const AccountDeletedPage = lazy(() =>
  import("@pages/account-deleted").then((m) => ({ default: m.AccountDeletedPage })),
);
const SessionExpiredPage = lazy(() =>
  import("@pages/session-expired").then((m) => ({ default: m.SessionExpiredPage })),
);
const BlockedPage = lazy(() => import("@pages/blocked").then((m) => ({ default: m.BlockedPage })));
const NotFoundPage = lazy(() =>
  import("@pages/not-found").then((m) => ({ default: m.NotFoundPage })),
);
const SettingsProfilePage = lazy(() =>
  import("@pages/settings-profile").then((m) => ({ default: m.SettingsProfilePage })),
);
const SettingsEditProfilePage = lazy(() =>
  import("@pages/settings-edit-profile").then((m) => ({ default: m.SettingsEditProfilePage })),
);
const SettingsSecurityPage = lazy(() =>
  import("@pages/settings-security").then((m) => ({ default: m.SettingsSecurityPage })),
);

function AppRoutes() {
  return (
    <Suspense fallback={null}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/verify" element={<VerifyPage />} />
        <Route path="/forgot" element={<ForgotPage />} />
        <Route path="/forgot/verify" element={<ForgotVerifyPage />} />
        <Route path="/reset" element={<ResetPage />} />
        <Route path="/cli/login" element={<CliLoginPage />} />
        <Route path="/account-deleted" element={<AccountDeletedPage />} />
        <Route path="/session-expired" element={<SessionExpiredPage />} />
        <Route path="/blocked" element={<BlockedPage />} />
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <SettingsProfilePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings/profile"
          element={
            <ProtectedRoute>
              <SettingsEditProfilePage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings/security"
          element={
            <ProtectedRoute>
              <SettingsSecurityPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/tasks"
          element={
            <ProtectedRoute>
              <TasksListPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/tasks/:taskId"
          element={
            <ProtectedRoute>
              <TaskDetailPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/sandbox/tasks/reports/:reportId"
          element={
            <ProtectedRoute>
              <TaskReportPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/sandbox/tasks"
          element={
            <ProtectedRoute>
              <SandboxPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/sandbox"
          element={
            <ProtectedRoute>
              <SandboxPage />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </Suspense>
  );
}

function AppWithToast() {
  const { toasts, removeToast } = useToast();

  return (
    <>
      <AppRoutes />
      <ToastContainer toasts={toasts} onRemove={removeToast} />
    </>
  );
}

export function App() {
  return (
    <AuthProvider>
      <ToastProvider>
        <AppWithToast />
      </ToastProvider>
    </AuthProvider>
  );
}
