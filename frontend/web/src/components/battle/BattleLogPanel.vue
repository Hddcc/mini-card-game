<script setup lang="ts">
/** 战斗日志（最新在上，最多 30 条；Vue 模板插值天然转义，修复旧版 raw 拼接隐患）。 */
import { computed } from 'vue'

import { MAX_BATTLE_LOGS } from '@/constants'
import type { BattleLog } from '@/types/battle'

const props = defineProps<{ logs: BattleLog[] }>()

const reversedLogs = computed(() => props.logs.slice(-MAX_BATTLE_LOGS).slice().reverse())
</script>

<template>
  <div class="h-32 overflow-y-auto border border-outline-variant bg-surface-container-high/60 p-3 thin-scrollbar">
    <p v-if="reversedLogs.length === 0" class="font-body-md text-sm text-on-surface-variant">战斗开始，等待行动…</p>
    <p v-for="(log, index) in reversedLogs" :key="index" class="mb-1 font-body-md text-sm text-on-surface-variant">
      第{{ log.round }}回合 · {{ log.text }}
    </p>
  </div>
</template>
