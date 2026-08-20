// Shared display mappings for status enums (mirrors backend domain/enums.go).
export const statusText = (s: number, kind: string): string => {
  const map: Record<string, Record<number, string>> = {
    generic: { 0: '停用', 1: '活跃' },
    field: { 0: '停用', 1: '可用', 2: '种植中', 3: '休耕' },
    plan: { 0: '已取消', 1: '计划中', 2: '已播种', 3: '生长中', 4: '已采收', 5: '已完成' },
    task: { 0: '已取消', 1: '计划中', 2: '进行中', 3: '已完成' },
    batch: { 0: '无效', 1: '有效', 2: '已耗尽', 3: '已过期' },
    produce: { 0: '已损耗', 1: '在库', 2: '已售', 3: '已加工' },
  }
  return map[kind]?.[s] ?? String(s)
}

export const statusTag = (s: number, kind: string): string => {
  const map: Record<string, Record<number, string>> = {
    generic: { 0: 'info', 1: 'success' },
    field: { 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' },
    plan: { 0: 'info', 1: '', 2: 'warning', 3: 'primary', 4: 'success', 5: 'success' },
    task: { 0: 'info', 1: '', 2: 'warning', 3: 'success' },
    batch: { 0: 'info', 1: 'success', 2: 'danger', 3: 'warning' },
    produce: { 0: 'danger', 1: 'success', 2: 'warning', 3: 'primary' },
  }
  return map[kind]?.[s] ?? 'info'
}

export const materialCategoryText: Record<number, string> = {
  1: '肥料', 2: '农药', 3: '种子', 4: '燃料', 5: '其他',
}

export const taskTypeText: Record<number, string> = {
  1: '施肥', 2: '灌溉', 3: '病虫害防治', 4: '采收', 5: '其他',
}

export const allocationTypeText: Record<number, string> = {
  0: '退回', 1: '领用', 2: '损耗',
}

export const allocationTypeTag: Record<number, string> = {
  0: 'success', 1: 'primary', 2: 'danger',
}
