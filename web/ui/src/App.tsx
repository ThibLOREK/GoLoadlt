import { Routes, Route, Navigate } from 'react-router-dom'
import ProjectsPage from './pages/ProjectsPage'
import EditorPage from './pages/EditorPage'
import ConnectionsPage from './pages/ConnectionsPage'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/projects" replace />} />
      <Route path="/projects" element={<ProjectsPage />} />
      <Route path="/projects/:projectId/edit" element={<EditorPage />} />
      <Route path="/connections" element={<ConnectionsPage />} />
    </Routes>
  )
}
