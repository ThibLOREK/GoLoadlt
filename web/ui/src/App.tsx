import { Routes, Route, Navigate } from "react-router-dom";
import { useEffect } from "react";
import { api } from "./api/client";
import Layout from "./components/Layout";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import PipelineList from "./pages/PipelineList";
import PipelineDesigner from "./pages/PipelineDesigner";
import RunHistory from "./pages/RunHistory";

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem("token");
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  useEffect(() => {
    const token = localStorage.getItem("token");
    if (token) api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  }, []);

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/*" element={
        <RequireAuth>
          <Layout>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/pipelines" element={<PipelineList />} />
              <Route path="/pipelines/:id/design" element={<PipelineDesigner />} />
              <Route path="/pipelines/:id/runs" element={<RunHistory />} />
            </Routes>
          </Layout>
        </RequireAuth>
      } />
    </Routes>
  );
}
