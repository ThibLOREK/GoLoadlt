import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from '@/components/Layout'
import ProjectsPage from '@/pages/ProjectsPage'
import EditorPage from '@/pages/EditorPage'
import ConnectionsPage from '@/pages/ConnectionsPage'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/projects" replace />} />
          <Route path="projects" element={<ProjectsPage />} />
          <Route path="projects/:projectId/edit" element={<EditorPage />} />
          <Route path="connections" element={<ConnectionsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
