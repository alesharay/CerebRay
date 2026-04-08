import { api } from './client'
import type { DashboardStats } from '../types'

export function getDashboardStats(): Promise<DashboardStats> {
  return api.get<DashboardStats>('/dashboard')
}
