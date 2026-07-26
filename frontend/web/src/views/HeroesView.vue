<script setup lang="ts">
/**
 * 英雄阵容（迁移自旧 mini_3）：
 * - 5 槽编辑：点槽选中 → 点未上阵神将入槽 → 自动跳到下一空槽
 * - 已上阵神将置灰不可点；槽位可清除
 * - 保存：过滤空槽、至少 1 人；成功/失败 toast
 * - 品质筛选与搜索在旧版是无事件的装饰，这里实现为真实功能（增强项）
 */
import { computed, onMounted, ref } from 'vue'

import { fetchTeam, saveTeam } from '@/api/team'
import HeroCard from '@/components/hero/HeroCard.vue'
import TeamSlot from '@/components/hero/TeamSlot.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import StoneButton from '@/components/ui/StoneButton.vue'
import { useToast } from '@/composables/useToast'
import { TEAM_SLOT_COUNT } from '@/constants'
import { qualityName } from '@/constants/quality'
import { useAssetStore } from '@/stores/assets'
import { useHeroStore } from '@/stores/hero'
import { usePlayerStore } from '@/stores/player'
import { errorText } from '@/types/api'
import type { HeroView } from '@/types/hero'

const toast = useToast()
const heroStore = useHeroStore()
const playerStore = usePlayerStore()
const assetStore = useAssetStore()

const activeTeam = ref<Array<number | null>>(Array.from({ length: TEAM_SLOT_COUNT }, () => null))
const selectedSlot = ref<number | null>(0)
const loading = ref(false)
const saving = ref(false)

// 增强：真实的品质筛选与搜索
const qualityFilter = ref<'all' | 5 | 4 | 3 | 2>('all')
const searchQuery = ref('')

const QUALITY_FILTERS: Array<{ value: 'all' | 5 | 4 | 3 | 2; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 5, label: 'SSR' },
  { value: 4, label: 'SR' },
  { value: 3, label: 'R' },
  { value: 2, label: 'N' },
]

const filteredHeroes = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return heroStore.heroes.filter((hero) => {
    if (qualityFilter.value !== 'all') {
      // N 档覆盖品质 1 与 2
      const matches = qualityFilter.value === 2 ? hero.quality <= 2 : hero.quality === qualityFilter.value
      if (!matches) return false
    }
    if (!query) return true
    const haystack = `${hero.name}${qualityName(hero.quality)}${hero.role}`.toLowerCase()
    return haystack.includes(query)
  })
})

const teamCount = computed(() => activeTeam.value.filter((id) => id !== null).length)

function heroInSlot(index: number): HeroView | null {
  const id = activeTeam.value[index]
  if (id === null || id === undefined) return null
  return heroStore.byId(id) ?? null
}

function isAssigned(hero: HeroView): boolean {
  return activeTeam.value.includes(hero.player_hero_id)
}

async function init(): Promise<void> {
  loading.value = true
  try {
    const [, team] = await Promise.all([
      heroStore.refresh(),
      fetchTeam(),
      playerStore.refresh(),
      assetStore.refresh(),
    ])
    const next = Array.from({ length: TEAM_SLOT_COUNT }, () => null) as Array<number | null>
    for (const entry of team) {
      if (entry.slot && entry.player_hero_id && entry.slot >= 1 && entry.slot <= TEAM_SLOT_COUNT) {
        next[entry.slot - 1] = entry.player_hero_id
      }
    }
    activeTeam.value = next
  } catch (error) {
    toast.show(errorText(error, '加载数据失败，请刷新重试'))
  } finally {
    loading.value = false
    selectedSlot.value = 0
  }
}

onMounted(init)

function selectSlot(index: number): void {
  selectedSlot.value = index
}

function clearSlot(index: number): void {
  activeTeam.value[index] = null
}

/** 点选神将入槽后自动跳到下一空槽（旧 handleHeroSelect 行为） */
function handleHeroSelect(hero: HeroView): void {
  let target = selectedSlot.value
  if (target === null) target = activeTeam.value.findIndex((id) => id === null)
  if (target === -1) target = 0
  activeTeam.value[target] = hero.player_hero_id
  const nextEmpty = activeTeam.value.findIndex((id) => id === null)
  if (nextEmpty !== -1) selectedSlot.value = nextEmpty
}

async function handleSave(): Promise<void> {
  const slots = activeTeam.value
    .map((heroId, index) => ({ slot: index + 1, player_hero_id: heroId ?? 0 }))
    .filter((entry) => entry.player_hero_id !== 0)
  if (slots.length === 0) {
    toast.show('请至少上阵 1 名英雄')
    return
  }
  saving.value = true
  try {
    await saveTeam({ slots })
    toast.show('保存成功')
    // 保存阵容会重算战力，同步刷新资料
    void playerStore.refresh().catch(() => undefined)
  } catch (error) {
    toast.show(`保存失败：${errorText(error)}`)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="mx-auto grid max-w-7xl grid-cols-1 gap-stack-lg lg:grid-cols-[380px_1fr]">
    <!-- 阵容面板 -->
    <section class="space-y-stack-md">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">groups</span>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-primary">出战阵容</h3>
        </div>
        <span class="rounded-full bg-surface-container px-3 py-1 font-stats-num text-label-sm text-on-surface-variant"
          >{{ teamCount }}/{{ TEAM_SLOT_COUNT }}</span
        >
      </div>

      <div class="grid grid-cols-2 gap-stack-md sm:grid-cols-3 lg:grid-cols-2">
        <TeamSlot
          v-for="(_, index) in activeTeam"
          :key="index"
          :slot-index="index"
          :hero="heroInSlot(index)"
          :selected="selectedSlot === index"
          @select="selectSlot(index)"
          @clear="clearSlot(index)"
        />
      </div>

      <StoneButton variant="primary" size="lg" block :disabled="saving" @click="handleSave">
        <span class="material-symbols-outlined" aria-hidden="true">{{ saving ? 'progress_activity' : 'save' }}</span>
        {{ saving ? '保存中…' : '保存阵容' }}
      </StoneButton>
      <p class="font-label-sm text-label-sm text-on-surface-variant">
        提示：保存后将按阵容重新计算战力，关卡挑战需先配置阵容。
      </p>
    </section>

    <!-- 神将背包 -->
    <section class="space-y-stack-md">
      <div class="flex flex-wrap items-center justify-between gap-stack-md">
        <div class="flex items-center gap-2">
          <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">style</span>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-primary">神将背包</h3>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button
            v-for="filter in QUALITY_FILTERS"
            :key="filter.value"
            class="rounded-full border px-3 py-1 font-label-sm text-label-sm transition-colors"
            :class="
              qualityFilter === filter.value
                ? 'border-primary-fixed bg-primary-fixed text-on-primary-fixed'
                : 'border-outline-variant text-on-surface-variant hover:text-on-surface'
            "
            @click="qualityFilter = filter.value"
          >
            {{ filter.label }}
          </button>
        </div>
      </div>

      <div class="group relative max-w-sm">
        <span
          class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant"
          aria-hidden="true"
          >search</span
        >
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索英雄名称、定位…"
          class="w-full border-b-2 border-outline-variant bg-surface-container-low py-2.5 pl-10 pr-4 text-on-surface outline-none transition-all placeholder:text-surface-variant focus:border-primary-container"
        />
      </div>

      <LoadingState v-if="loading && heroStore.heroes.length === 0" text="神将加载中…" />
      <EmptyState
        v-else-if="heroStore.heroes.length === 0"
        icon="auto_awesome"
        title="暂无英雄"
        description="请先前往召唤获得英雄。"
      >
        <router-link
          :to="{ name: 'gacha' }"
          class="border-2 border-primary-fixed bg-ink-wash px-6 py-2 font-title-md text-label-sm text-primary-fixed transition-all hover:bg-primary-fixed hover:text-on-primary-fixed"
        >
          前往召唤
        </router-link>
      </EmptyState>
      <EmptyState
        v-else-if="filteredHeroes.length === 0"
        icon="search_off"
        title="没有符合条件的神将"
        description="调整筛选或搜索条件试试。"
      />
      <div v-else class="grid grid-cols-2 gap-stack-md sm:grid-cols-3 xl:grid-cols-4">
        <HeroCard
          v-for="hero in filteredHeroes"
          :key="hero.player_hero_id"
          :hero="hero"
          :disabled="isAssigned(hero)"
          @select="handleHeroSelect"
        />
      </div>
    </section>
  </div>
</template>
