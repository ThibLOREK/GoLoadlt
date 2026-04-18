import axios from 'axios'
import type { Project, Connection } from '@/types/api'

const api = axios.create({ baseURL: '/api/v1' })

// --- Projets ---
export const listProjects = () =>
  api.get<Project[]>('/projects').then(r => r.data ?? [])

export const getProject = (id: string) =>
  api.get<Project>(`/projects/${id}`).then(r => r.data)

export const createProject = (p: Partial<Project>) =>
  api.post<Project>('/projects', p).then(r => r.data)

export const updateProject = (id: string, p: Project) =>
  api.put<Project>(`/projects/${id}`, p).then(r => r.data)

export const deleteProject = (id: string) =>
  api.delete(`/projects/${id}`)

export const runProject = (id: string) =>
  api.post(`/projects/${id}/run`).then(r => r.data)

export const exportXML = (id: string) =>
  api.get(`/projects/${id}/xml`, { responseType: 'blob' }).then(r => r.data)

// --- Catalogue ---
export const getCatalogue = () =>
  api.get<any[]>('/catalogue').then(r => r.data ?? [])

// --- Connexions ---
export const listConnections = () =>
  api.get<Connection[]>('/connections').then(r => r.data ?? [])

export const getConnection = (id: string) =>
  api.get<Connection>(`/connections/${id}`).then(r => r.data)

export const createConnection = (c: Partial<Connection>) =>
  api.post<Connection>('/connections', c).then(r => r.data)

export const updateConnection = (id: string, c: Connection) =>
  api.put<Connection>(`/connections/${id}`, c).then(r => r.data)

export const deleteConnection = (id: string) =>
  api.delete(`/connections/${id}`)

export const testConnection = (id: string) =>
  api.post(`/connections/${id}/test`).then(r => r.data)

// --- Environnement ---
export const getEnvironment = () =>
  api.get('/environment').then(r => r.data)

export const switchEnvironment = (env: string) =>
  api.put('/environment', { env }).then(r => r.data)
