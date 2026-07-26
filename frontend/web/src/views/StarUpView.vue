<script setup lang="ts">
/**
 * 碎片升星（迁移自旧 mini_7）：
 * - 分阶段加载：先神将（失败→整卡错误态，区分超时文案），后资产（失败仅行内提示）
 * - 搜索（名称/品质名/定位）；10 星行；材料三态徽章 + 按钮客户端预校验
 * - 升星成功：单条替换神将 + PascalCase 资产归一化写入全局资产栏
 *   （顺带修复旧页直接读 PascalCase 字段的 gold 显示 bug）
 *
 * 布局：反转旧的「360px 列表 + 右详情」主从结构 —— 升星详情横幅置顶
 * （立绘氛围层 + 身份/星级 + 材料/确认），名册改全宽卡片网格铺满下方。
 */
import { computed, onMounted, ref } from 'vue'

import { starUpHero } from '@/api/hero'
import { normalizeAssets } from '@/api/normalizers'
import AppInput from '@/components/ui/AppInput.vue'
import JadePanel from '@/components/ui/JadePanel.vue'
import JadeProgressBar from '@/components/ui/JadeProgressBar.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import RarityBadge from '@/components/ui/RarityBadge.vue'
import { MAX_HERO_STAR } from '@/constants'
import { HERO_IMAGES, heroImage } from '@/constants/heroAssets'
import { qualityName, qualityTextClass } from '@/constants/quality'
import { useAssetStore } from '@/stores/assets'
import { useHeroStore } from '@/stores/hero'
import { ApiError, errorText } from '@/types/api'
import type { HeroView, StarUpCost } from '@/types/hero'

const heroStore = useHeroStore()
const assetStore = useAssetStore()

const selectedId = ref(0)
const heroQuery = ref('')
const loadError = ref('')
const assetMessage = ref('')
const starUpPending = ref(false)
const message = ref<{ text: string; tone: 'neutral' | 'success' | 'error' } | null>(null)

const heroes = computed(() => heroStore.heroes)

const selectedHero = computed<HeroView | null>(
  () => heroes.value.find((hero) => hero.player_hero_id === selectedId.value) ?? heroes.value[0] ?? null,
)

const filteredHeroes = computed(() => {
  const keyword = heroQuery.value.trim().toLowerCase()
  if (!keyword) return heroes.value
  return heroes.value.filter((hero) => {
    const text = `${hero.name || ''} ${qualityName(hero.quality)} ${hero.role || ''}`.toLowerCase()
    return text.includes(keyword)
  })
})

function costOf(hero: HeroView | null): StarUpCost {
  return hero?.next_star_cost ?? { shard: 0, gold: 0 }
}

function fmt(value: number): string {
  return Number(value || 0).toLocaleString('zh-CN')
}

const materials = computed(() => {
  const hero = selectedHero.value
  if (!hero) return null
  const cost = costOf(hero)
  const maxStar = Boolean(hero.max_star)
  const shardOk = maxStar || Number(hero.shard || 0) >= Number(cost.shard || 0)
  const goldOk = maxStar || assetStore.gold >= Number(cost.gold || 0)
  return {
    cost,
    maxStar,
    shardOk,
    goldOk,
    canStarUp: !maxStar && shardOk && goldOk,
  }
})

function badgeText(ok: boolean, maxStar: boolean): string {
  return maxStar ? '已满星' : ok ? '充足' : '不足'
}

/** 名册头部的品质分布 chips（现有数据的汇总展示） */
const qualityCounts = computed(() => {
  const counts = new Map<string, { name: string; quality: number; count: number }>()
  for (const hero of heroes.value) {
    const name = qualityName(hero.quality)
    const entry = counts.get(name)
    if (entry) entry.count += 1
    else counts.set(name, { name, quality: hero.quality, count: 1 })
  }
  return [...counts.values()].sort((a, b) => b.quality - a.quality)
})

/** 分阶段加载：神将失败直接进入错误态；资产失败只提示不阻塞（旧 mini_7 行为） */
async function loadData(): Promise<void> {
  loadError.value = ''
  assetMessage.value = ''
  try {
    await heroStore.refresh()
    if (heroes.value.length && !heroes.value.some((hero) => hero.player_hero_id === selectedId.value)) {
      selectedId.value = heroes.value[0]!.player_hero_id
    }
    if (!heroes.value.length) selectedId.value = 0
  } catch (error) {
    loadError.value =
      error instanceof ApiError && error.kind === 'timeout'
        ? '神将接口请求超时，请刷新重试'
        : '神将加载失败，请刷新重试'
    return
  }
  try {
    await assetStore.refresh()
  } catch (error) {
    assetMessage.value =
      error instanceof ApiError && error.kind === 'timeout' ? '资产接口请求超时' : '资产加载失败，请刷新重试'
  }
}

onMounted(loadData)

async function handleStarUp(): Promise<void> {
  const hero = selectedHero.value
  if (!hero || starUpPending.value) return
  starUpPending.value = true
  message.value = { text: '升星中...', tone: 'neutral' }
  try {
    const output = await starUpHero({ player_hero_id: hero.player_hero_id })
    heroStore.replaceHero(output.hero)
    // star-up 返回的 assets 是 PascalCase（model.PlayerAsset 无 json tag），归一化后同步全局资产栏
    assetStore.apply(normalizeAssets(output.assets, assetStore.assets))
    message.value = { text: '升星成功', tone: 'success' }
  } catch (error) {
    message.value = { text: errorText(error, '升星失败'), tone: 'error' }
  } finally {
    starUpPending.value = false
  }
}
</script>

<template>
  <div class="mx-auto flex max-w-7xl flex-col gap-stack-lg">
    <!-- ① 升星横幅（详情置顶；无神将时原 EmptyState 占位，条件与旧版一致） -->
    <EmptyState
      v-if="!heroStore.loading && !loadError && !heroes.length"
      icon="auto_awesome"
      title="暂无神将"
      description="前往召唤获得神将后即可升星。"
    >
      <router-link
        :to="{ name: 'gacha' }"
        class="border-2 border-primary-fixed bg-ink-wash px-6 py-2 font-title-md text-label-sm text-primary-fixed transition-all hover:bg-primary-fixed hover:text-on-primary-fixed"
      >
        前往召唤
      </router-link>
    </EmptyState>

    <JadePanel v-else-if="selectedHero && materials" tone="gold" class="overflow-hidden">
      <!-- 氛围层：选中神将立绘虚化铺底 -->
      <img
        :src="heroImage(selectedHero.hero_config_id) ?? HERO_IMAGES[1]"
        alt=""
        aria-hidden="true"
        class="pointer-events-none absolute inset-0 h-full w-full object-cover object-top opacity-25 blur-sm"
      />
      <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-black/85 via-ink-wash/55 to-black/80"></div>
      <div class="ink-overlay absolute inset-0"></div>

      <div
        class="relative z-10 grid gap-stack-lg p-6 md:grid-cols-[240px_1fr] md:p-10 xl:grid-cols-[280px_minmax(0,1fr)_320px]"
      >
        <!-- A：立绘 -->
        <div class="relative overflow-hidden border-2 border-outline-variant bg-surface-container-low">
          <img
            :src="heroImage(selectedHero.hero_config_id) ?? HERO_IMAGES[1]"
            :alt="selectedHero.name"
            class="aspect-[3/4] w-full object-cover"
          />
          <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black p-stack-md">
            <p class="font-display-hero text-title-md text-on-surface">{{ selectedHero.name }}</p>
            <div class="mt-1 flex flex-wrap items-center gap-2">
              <RarityBadge :quality="selectedHero.quality" />
              <span class="font-label-sm text-label-sm text-on-surface-variant"
                >{{ selectedHero.role || '未知定位' }} · 等级 {{ selectedHero.level || 1 }}</span
              >
            </div>
          </div>
        </div>

        <!-- B：身份 + 星级 -->
        <div class="flex flex-col justify-center gap-stack-md">
          <h2 class="glow-gold font-display-hero text-headline-lg text-primary-fixed md:text-display-hero">
            {{ selectedHero.name }}
          </h2>
          <div class="flex flex-wrap items-center gap-2">
            <RarityBadge :quality="selectedHero.quality" />
            <span
              class="rounded-full border border-outline-variant bg-surface-container-lowest/70 px-3 py-0.5 font-label-sm text-label-sm text-on-surface-variant"
              >{{ selectedHero.role || '未知定位' }}</span
            >
            <span
              class="rounded-full border border-outline-variant bg-surface-container-lowest/70 px-3 py-0.5 font-label-sm text-label-sm text-on-surface-variant"
              >等级 {{ selectedHero.level || 1 }}</span
            >
          </div>

          <div class="rounded-lg border border-outline-variant bg-surface-container/70 p-stack-lg backdrop-blur-sm">
            <div class="flex items-center justify-between gap-stack-md">
              <div>
                <p class="font-label-sm text-label-sm uppercase text-on-surface-variant">星级</p>
                <p class="mt-1 font-stats-num text-stats-num text-on-surface">
                  {{ selectedHero.star || 1 }} 星
                  <span class="text-on-surface-variant">→</span>
                  <span class="text-primary-fixed">{{
                    materials.maxStar ? selectedHero.star : Number(selectedHero.star || 1) + 1
                  }}</span>
                  星
                </p>
              </div>
              <p
                class="font-title-md text-headline-lg-mobile"
                :class="materials.canStarUp || materials.maxStar ? 'text-tertiary-container' : 'text-error'"
              >
                {{ materials.maxStar ? '满星' : materials.canStarUp ? '可升星' : '材料不足' }}
              </p>
            </div>
            <div class="mt-stack-md flex flex-wrap gap-1" aria-label="星级进度">
              <span
                v-for="starIndex in MAX_HERO_STAR"
                :key="starIndex"
                class="material-symbols-outlined"
                :class="
                  starIndex <= Number(selectedHero.star || 0)
                    ? 'icon-filled text-quality-ssr'
                    : 'text-outline-variant'
                "
                aria-hidden="true"
                >star</span
              >
            </div>
          </div>
        </div>

        <!-- C：材料 + 确认（md 换行跨两列，xl 归右列） -->
        <div class="flex flex-col gap-stack-md md:col-span-2 xl:col-span-1">
          <div class="grid grid-cols-1 gap-gutter sm:grid-cols-2 xl:grid-cols-1">
            <div class="rounded-lg border border-outline-variant bg-surface-container/70 p-stack-lg backdrop-blur-sm">
              <div class="flex items-center justify-between">
                <span class="font-label-sm text-label-sm text-on-surface-variant">神将碎片</span>
                <span
                  class="font-label-sm text-label-sm"
                  :class="materials.shardOk ? 'text-tertiary-container' : 'text-error'"
                  >{{ badgeText(materials.shardOk, materials.maxStar) }}</span
                >
              </div>
              <p
                class="mt-2 font-stats-num text-stats-num"
                :class="materials.shardOk ? 'text-tertiary-container' : 'text-error'"
              >
                {{ fmt(selectedHero.shard) }}
              </p>
              <p class="font-label-sm text-label-sm text-on-surface-variant">
                需要 {{ materials.maxStar ? '已满星' : `${fmt(materials.cost.shard)} 个` }}
              </p>
              <div class="mt-stack-md space-y-1">
                <JadeProgressBar
                  :value="materials.maxStar ? 1 : Number(selectedHero.shard || 0)"
                  :max="materials.maxStar ? 1 : Math.max(1, Number(materials.cost.shard || 0))"
                  height-class="h-2"
                />
                <p class="text-right font-label-sm text-[10px] text-on-surface-variant">
                  {{ materials.maxStar ? '满星' : `${fmt(selectedHero.shard)} / ${fmt(materials.cost.shard)}` }}
                </p>
              </div>
            </div>

            <div class="rounded-lg border border-outline-variant bg-surface-container/70 p-stack-lg backdrop-blur-sm">
              <div class="flex items-center justify-between">
                <span class="font-label-sm text-label-sm text-on-surface-variant">金币</span>
                <span
                  class="font-label-sm text-label-sm"
                  :class="materials.goldOk ? 'text-tertiary-container' : 'text-error'"
                  >{{ badgeText(materials.goldOk, materials.maxStar) }}</span
                >
              </div>
              <p
                class="mt-2 font-stats-num text-stats-num"
                :class="materials.goldOk ? 'text-tertiary-container' : 'text-error'"
              >
                {{ fmt(assetStore.gold) }}
              </p>
              <p class="font-label-sm text-label-sm text-on-surface-variant">
                需要 {{ materials.maxStar ? '已满星' : fmt(materials.cost.gold) }}
              </p>
              <p v-if="assetMessage" class="mt-stack-md font-label-sm text-label-sm text-error">
                {{ assetMessage }}
              </p>
            </div>
          </div>

          <button
            type="button"
            class="focus-ring flex min-h-[56px] w-full cursor-pointer items-center justify-center gap-2 bg-primary-fixed px-6 py-4 font-title-md text-title-md text-on-primary-fixed transition-colors duration-200 hover:bg-primary-fixed-dim disabled:cursor-not-allowed disabled:border disabled:border-outline-variant disabled:bg-surface-container-high disabled:text-on-surface-variant"
            :class="materials.canStarUp ? 'divine-glow' : ''"
            :disabled="starUpPending || materials.maxStar || !materials.shardOk || !materials.goldOk"
            @click="handleStarUp"
          >
            <span class="material-symbols-outlined icon-filled" aria-hidden="true">{{
              starUpPending ? 'progress_activity' : 'upgrade'
            }}</span>
            <span>{{ starUpPending ? '升星中…' : '确认升星' }}</span>
          </button>
          <p
            aria-live="polite"
            class="min-h-[24px] text-center text-sm"
            :class="
              message?.tone === 'success'
                ? 'text-tertiary-container'
                : message?.tone === 'error'
                  ? 'text-error'
                  : 'text-on-surface-variant'
            "
          >
            {{ message?.text ?? '' }}
          </p>
        </div>
      </div>
    </JadePanel>

    <!-- ② 名册（全宽网格） -->
    <section class="space-y-stack-md">
      <div class="flex flex-wrap items-center justify-between gap-stack-md">
        <div class="flex flex-wrap items-center gap-stack-md">
          <div class="flex items-center gap-2">
            <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">stars</span>
            <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-primary">已拥有神将</h3>
          </div>
          <span class="rounded-full bg-surface-container px-3 py-1 font-label-sm text-label-sm text-on-surface-variant">
            {{ loadError ? '加载失败' : heroStore.loading ? '加载中' : `${heroes.length} 位` }}
          </span>
          <div v-if="qualityCounts.length" class="hidden items-center gap-2 sm:flex">
            <span
              v-for="entry in qualityCounts"
              :key="entry.name"
              class="font-label-sm text-label-sm"
              :class="qualityTextClass(entry.quality)"
              >{{ entry.name }} {{ entry.count }}</span
            >
          </div>
        </div>
        <AppInput v-model="heroQuery" icon="search" placeholder="搜索名称、品质、定位…" class="w-full sm:w-72" />
      </div>

      <div
        v-if="heroStore.loading"
        class="border border-outline-variant bg-surface-container-low p-4 text-sm text-on-surface-variant"
      >
        神将加载中...
      </div>
      <div v-else-if="loadError" class="border border-error/60 bg-surface-container-low p-4 text-sm text-error">
        {{ loadError }}
        <button class="ml-2 underline" @click="loadData">重试</button>
      </div>
      <div
        v-else-if="!heroes.length"
        class="border border-outline-variant bg-surface-container-low p-4 text-sm text-on-surface-variant"
      >
        暂无已拥有神将。
      </div>
      <div
        v-else-if="!filteredHeroes.length"
        class="border border-outline-variant bg-surface-container-low p-4 text-sm text-on-surface-variant"
      >
        没有匹配的神将。
      </div>
      <div v-else class="grid grid-cols-2 gap-gutter sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        <button
          v-for="hero in filteredHeroes"
          :key="hero.player_hero_id"
          type="button"
          class="focus-ring group relative cursor-pointer overflow-hidden border-2 bg-surface-container-low text-left transition-all duration-200 hover:border-primary-fixed"
          :class="
            hero.player_hero_id === selectedHero?.player_hero_id
              ? 'border-primary-fixed shadow-[0_0_15px_rgba(255,215,0,0.4)]'
              : 'border-outline-variant'
          "
          :aria-pressed="hero.player_hero_id === selectedHero?.player_hero_id"
          @click="selectedId = hero.player_hero_id"
        >
          <img
            :src="heroImage(hero.hero_config_id) ?? HERO_IMAGES[1]"
            :alt="`${hero.name || '神将'}立绘`"
            loading="lazy"
            class="aspect-[3/4] w-full object-cover transition-transform duration-300 group-hover:scale-105"
          />
          <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black via-black/70 to-transparent p-stack-md pt-8">
            <div class="flex items-center justify-between gap-2">
              <p class="truncate font-title-md text-body-md text-on-surface">{{ hero.name || '神将' }}</p>
              <span class="whitespace-nowrap font-label-sm text-label-sm text-on-surface-variant"
                >{{ hero.star || 1 }}星</span
              >
            </div>
            <div class="mt-1.5 flex items-center gap-2">
              <RarityBadge :quality="hero.quality" />
              <div class="min-w-0 flex-1 space-y-0.5">
                <JadeProgressBar
                  v-if="!hero.max_star"
                  :value="Number(hero.shard || 0)"
                  :max="Math.max(1, Number(hero.next_star_cost?.shard || 1))"
                  variant="gold"
                  height-class="h-1"
                />
                <p class="truncate text-right font-label-sm text-[10px] text-on-surface-variant">
                  {{ hero.max_star ? '已满星' : `碎片 ${fmt(hero.shard)} / ${fmt(hero.next_star_cost?.shard || 0)}` }}
                </p>
              </div>
            </div>
          </div>
        </button>
      </div>
    </section>
  </div>
</template>
