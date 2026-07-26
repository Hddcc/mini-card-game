import { http } from './http'
import type { TaskView } from '@/types/task'

/** data 为裸数组，首次调用会自动创建当天任务行 */
export function fetchDailyTasks(): Promise<TaskView[]> {
  return http.get<TaskView[]>('/api/v1/tasks/daily')
}

/** 成功时 data 为奖励数组，前端只关心 code===0 后刷新任务与资产 */
export function claimTask(taskId: number): Promise<unknown> {
  return http.post<unknown>('/api/v1/tasks/claim', { task_id: taskId })
}
