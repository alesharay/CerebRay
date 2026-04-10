import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

vi.mock('../api/auth', () => ({
  getMe: vi.fn(),
  logout: vi.fn(),
}))

import { getMe, logout as apiLogout } from '../api/auth'

const mockedGetMe = vi.mocked(getMe)
const mockedLogout = vi.mocked(apiLogout)

describe('authStore', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: null, loading: true })
    vi.clearAllMocks()
  })

  it('starts with no user and loading true', () => {
    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.loading).toBe(true)
  })

  it('sets user on successful checkAuth', async () => {
    const fakeUser = { id: 1, email: 'test@test.com', name: 'Test', oidc_subject: 's', avatar_url: '', created_at: '', updated_at: '' }
    mockedGetMe.mockResolvedValueOnce(fakeUser)

    await useAuthStore.getState().checkAuth()

    const state = useAuthStore.getState()
    expect(state.user).toEqual(fakeUser)
    expect(state.loading).toBe(false)
  })

  it('clears user on failed checkAuth', async () => {
    mockedGetMe.mockRejectedValueOnce(new Error('401'))

    await useAuthStore.getState().checkAuth()

    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.loading).toBe(false)
  })

  it('clears user on logout', async () => {
    useAuthStore.setState({ user: { id: 1, email: 'a@b.com', name: 'A', oidc_subject: 's', avatar_url: '', created_at: '', updated_at: '' } })
    mockedLogout.mockResolvedValueOnce(undefined)

    await useAuthStore.getState().logout()

    expect(useAuthStore.getState().user).toBeNull()
  })
})
