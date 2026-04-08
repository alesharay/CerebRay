import { api } from './client'
import type { AIUsage } from '../types'

export function getUsage(): Promise<AIUsage> {
  return api.get<AIUsage>('/settings/usage')
}
