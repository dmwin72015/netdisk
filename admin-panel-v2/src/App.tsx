import { lazy, Suspense } from "react";
import { createBrowserRouter, Navigate } from "react-router-dom";
import BasicLayout from "@/layouts/BasicLayout";
import ProtectedRoute from "@/layouts/ProtectedRoute";

const Login = lazy(() => import("@/features/login/pages/Login"));
const Workplace = lazy(() => import("@/features/dashboard/pages/Workplace"));
const UserManagement = lazy(
  () => import("@/features/users/pages/UserManagement"),
);
const FilesManagement = lazy(
  () => import("@/features/files/pages/FilesManagement"),
);
const PhysicalFiles = lazy(
  () => import("@/features/physical-files/pages/PhysicalFiles"),
);
const StorageAnalysis = lazy(
  () => import("@/features/storage/pages/StorageAnalysis"),
);
const ActivityLogsManagement = lazy(
  () => import("@/features/activity-logs/pages/ActivityLogsManagement"),
);
const CleanupManagement = lazy(
  () => import("@/features/cleanup/pages/CleanupManagement"),
);
const SystemSettings = lazy(
  () => import("@/features/settings/pages/SystemSettings"),
);

const Dashboard = lazy(() => import("@/demos/dashboard"));
const Exception403 = lazy(() => import("@/demos/exception/403"));
const Exception404 = lazy(() => import("@/demos/exception/404"));
const Exception500 = lazy(() => import("@/demos/exception/500"));
const FormBasic = lazy(() => import("@/demos/form/basic"));
const ProfileBasic = lazy(() => import("@/demos/profile/basic"));
const ResultSuccess = lazy(() => import("@/demos/result/success"));
const Chatbot = lazy(() => import("@/demos/chatbot"));
const Register = lazy(() => import("@/demos/register"));
const RegisterResult = lazy(() => import("@/demos/register-result"));

function Loader() {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-lg">Loading...</div>
    </div>
  );
}

export const router = createBrowserRouter([
  {
    path: "/login",
    element: (
      <Suspense fallback={<Loader />}>
        <Login />
      </Suspense>
    ),
  },
  {
    path: "/",
    element: (
      <ProtectedRoute>
        <BasicLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Navigate to="/admin/dashboard" replace /> },
      {
        path: "admin/dashboard",
        element: (
          <Suspense fallback={<Loader />}>
            <Workplace />
          </Suspense>
        ),
      },
      {
        path: "admin/users",
        element: (
          <Suspense fallback={<Loader />}>
            <UserManagement />
          </Suspense>
        ),
      },
      {
        path: "admin/files",
        element: (
          <Suspense fallback={<Loader />}>
            <FilesManagement />
          </Suspense>
        ),
      },
      {
        path: "admin/storage",
        element: (
          <Suspense fallback={<Loader />}>
            <StorageAnalysis />
          </Suspense>
        ),
      },
      {
        path: "admin/physical-files",
        element: (
          <Suspense fallback={<Loader />}>
            <PhysicalFiles />
          </Suspense>
        ),
      },
      {
        path: "admin/activity-logs",
        element: (
          <Suspense fallback={<Loader />}>
            <ActivityLogsManagement />
          </Suspense>
        ),
      },
      {
        path: "admin/cleanup",
        element: (
          <Suspense fallback={<Loader />}>
            <CleanupManagement />
          </Suspense>
        ),
      },
      {
        path: "admin/settings",
        element: (
          <Suspense fallback={<Loader />}>
            <SystemSettings />
          </Suspense>
        ),
      },
      {
        path: "welcome",
        element: (
          <Suspense fallback={<Loader />}>
            <Dashboard />
          </Suspense>
        ),
      },
      { path: "account/*", element: <FormBasic /> },
      { path: "exception/403", element: <Exception403 /> },
      { path: "exception/404", element: <Exception404 /> },
      { path: "exception/500", element: <Exception500 /> },
      { path: "form/*", element: <FormBasic /> },
      { path: "profile/*", element: <ProfileBasic /> },
      { path: "result/*", element: <ResultSuccess /> },
      { path: "chatbot", element: <Chatbot /> },
      { path: "register", element: <Register /> },
      { path: "register-result", element: <RegisterResult /> },
    ],
  },
]);
