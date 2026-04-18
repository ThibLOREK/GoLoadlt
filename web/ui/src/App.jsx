import React from 'react'
import { Routes, Route, Navigate, Link, useNavigate } from 'react-router-dom'
import Login from './pages/Login'
import Pipelines from './pages/Pipelines'
import Runs from './pages/Runs'
import PipelineConfig from './pages/PipelineConfig'

const s = {
  nav: { display:'flex', alignItems:'center', gap:'1.5rem', padding:'1rem 2rem',
         background:'#1e293b', borderBottom:'1px solid #334155' },
  logo: { fontWeight:700, fontSize:'1.2rem', color:'#38bdf8', textDecoration:'none' },
  link: { color:'#94a3b8', textDecoration:'none', fontSize:'.9rem' },
  btn:  { marginLeft:'auto', background:'#ef4444', border:'none', color:'#fff',
          padding:'.4rem .9rem', borderRadius:6, cursor:'pointer', fontSize:'.85rem' },
}

function Nav() {
  const nav = useNavigate()
  function logout() { localStorage.removeItem('token'); nav('/login') }
  return (
    <nav style={s.nav}>
      <Link to="/pipelines" style={s.logo}>⚡ GoLoadIt</Link>
      <Link to="/pipelines" style={s.link}>Pipelines</Link>
      <button onClick={logout} style={s.btn}>Déconnexion</button>
    </nav>
  )
}

function Protected({ children }) {
  return localStorage.getItem('token') ? children : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/pipelines" element={<Protected><Nav /><Pipelines /></Protected>} />
      <Route path="/pipelines/:id/runs" element={<Protected><Nav /><Runs /></Protected>} />
      <Route path="/pipelines/:id/config" element={<Protected><Nav /><PipelineConfig /></Protected>} />
      <Route path="*" element={<Navigate to="/pipelines" replace />} />
    </Routes>
  )
}