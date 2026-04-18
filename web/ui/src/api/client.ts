import axios from 'axios'
import type { Project, Connection, BlockMeta, RunReport } from '@/types/api'

const http = axios.create({ baseURL: '/api/v1' })

// ── Projets ──────────────────────────────────────────────
export const listProjects  = () => http.get<Project[]>('/projects').then(r => r.data)
export const getProject    = (id: string) => http.get<Project>(`/projects/${id}`).then(r => r.data)
export const createProject = (p: Partial<Project>) => http.post<Project>('/projects', p).then(r => r.data)
export const updateProject = (id: string, p: Partial<Project>) => http.put<Project>(`/projects/${id}`, p).then(r => r.data)
export const deleteProject = (id: string) => http.delete(`/projects/${id}`)
export const runProject    = (id: string) => http.post<RunReport>(`/projects/${id}/run`).then(r => r.data)
export const exportXML     = (id: string) => http.get<string>(`/projects/${id}/xml`, { responseType: 'text' }).then(r => r.data)

// ── Catalogue ────────────────────────────────────────────
export const getCatalogue = () => http.get<BlockMeta[]>('/catalogue').then(r => r.data)

// ── Connexions ───────────────────────────────────────────
export const listConnections  = () => http.get<Connection[]>('/connections').then(r => r.data)
export const createConnection = (c: Partial<Connection>) => http.post<Connection>('/connections', c).then(r => r.data)
export const updateConnection = (id: string, c: Partial<Connection>) => http.put<Connection>(`/connections/${id}`, c).then(r => r.data)
export const deleteConnection = (id: string) => http.delete(`/connections/${id}`)
export const testConnection   = (id: string) => http.post(`/connections/${id}/test`).then(r => r.data)

// ── Environnement ─────────────────────────────────────────
export const getEnv    = () => http.get<{ activeEnv: string }>('/environment').then(r => r.data)
export const switchEnv = (env: string) => http.put('/environment', { env }).then(r => r.data)
