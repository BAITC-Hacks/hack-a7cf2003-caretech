const API_PREFIX = '/api'

async function request(path, options = {}) {
  const response = await fetch(`${API_PREFIX}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })

  let payload
  try {
    payload = await response.json()
  } catch {
    payload = null
  }

  if (!response.ok) {
    const error = new Error(payload?.error || `API request failed with ${response.status}`)
    error.status = response.status
    error.payload = payload
    throw error
  }

  return payload
}

export const api = {
  search: (query = '') => request(`/search?q=${encodeURIComponent(query)}`),
  product: (id) => request(`/products/${encodeURIComponent(id)}`),
  alternatives: (id) => request(`/products/${encodeURIComponent(id)}/alternatives`),
  chat: (body) => request('/chat', { method: 'POST', body: JSON.stringify(body) }),
  confirm: (token) => request('/cart/confirm', { method: 'POST', body: JSON.stringify({ confirmation_token: token }) }),
}
