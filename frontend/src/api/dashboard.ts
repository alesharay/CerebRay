import { api } from './client'
import type { DashboardStats, AnalyticsData } from '../types'

export function getDashboardStats(): Promise<DashboardStats> {
  return api.get<DashboardStats>('/dashboard')
}

export function getAnalytics(staleDays = 14): Promise<AnalyticsData> {
  return api.get<AnalyticsData>(`/dashboard/analytics?stale_days=${staleDays}`)
}
