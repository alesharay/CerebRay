import type { User } from '../types'

export async function getMe(): Promise<User> {
  const res = await fetch('/api/v1/auth/me', { credentials: 'include' })
  if (!res.ok) throw new Error('Not authenticated')
  return res.json()
}

export function login() {
  window.location.href = '/auth/login'
}

export async function logout() {
  await fetch('/auth/logout', { method: 'POST', credentials: 'include' })
  window.location.href = '/'
}
