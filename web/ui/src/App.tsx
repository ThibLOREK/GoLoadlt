import { Routes, Route } from "react-router-dom";
import Layout from "./components/Layout";
import PipelineList from "./pages/PipelineList";
import PipelineDesigner from "./pages/PipelineDesigner";
import RunHistory from "./pages/RunHistory";

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<PipelineList />} />
        <Route path="/pipelines/:id/design" element={<PipelineDesigner />} />
        <Route path="/pipelines/:id/runs" element={<RunHistory />} />
      </Routes>
    </Layout>
  );
}
