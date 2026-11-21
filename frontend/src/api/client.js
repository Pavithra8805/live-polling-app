const API_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '')

async function request(path, options = {}) {
  const token = localStorage.getItem('poll_token')
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  })
  if (response.status === 204) return null

  const data = await response.json().catch(() => null)
  if (!response.ok) {
    const message = data?.error?.message || 'Something went wrong'
    const error = new Error(message)
    error.status = response.status
    error.code = data?.error?.code
    throw error
  }
  return data
}

export const api = {
  signup: (email, password) => request('/api/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  }),
  login: (email, password) => request('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  }),
  createPoll: (payload) => request('/api/polls', {
    method: 'POST',
    body: JSON.stringify(payload),
  }),
  listMine: () => request('/api/polls/mine'),
  getPoll: (slug) => request(`/api/polls/${encodeURIComponent(slug)}`),
  getResults: (slug) => request(`/api/polls/${encodeURIComponent(slug)}/results`),
  vote: (slug, optionId) => request(`/api/polls/${encodeURIComponent(slug)}/vote`, {
    method: 'POST',
    body: JSON.stringify({ optionId }),
  }),
  closePoll: (slug) => request(`/api/polls/${encodeURIComponent(slug)}/close`, { method: 'PATCH' }),
  deletePoll: (slug) => request(`/api/polls/${encodeURIComponent(slug)}`, { method: 'DELETE' }),
}

export const streamUrl = (slug) =>
  `${API_URL}/api/polls/${encodeURIComponent(slug)}/stream`
