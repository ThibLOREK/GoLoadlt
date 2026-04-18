const BASE = '/api/v1'

function token() {
  return localStorage.getItem('token')
}

function headers() {
  return {
    'Content-Type': 'application/json',
    ...(token() ? { Authorization: `Bearer ${token()}` } : {}),
  }
}

async function req(method, path, body) {
  const res = await fetch(BASE + path, {
    method,
    headers: headers(),
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
  }
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || res.statusText)
  return data
}

export const api = {
  login:           (email, password)  => req('POST', '/auth/login', { email, password }),
  register:        (email, password)  => req('POST', '/auth/register', { email, password }),
  listPipelines:   ()                 => req('GET',  '/pipelines'),
  createPipeline:  (body)             => req('POST', '/pipelines', body),
  deletePipeline:  (id)               => req('DELETE', `/pipelines/${id}`),
  runPipeline:     (id)               => req('POST', `/pipelines/${id}/runs`),
  listRuns:        (id)               => req('GET',  `/pipelines/${id}/runs`),
  getPipeline:     (id)               => req('GET',    `/pipelines/${id}`),
  updatePipeline:  (id, body)         => req('PUT',    `/pipelines/${id}`, body),
}