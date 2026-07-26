<script setup lang="ts">
/**
 * 大厅（迁移自旧 mini_5）：
 * - 资料卡：本地头像上传、行内改名（Enter/Escape、服务端中文错误行内展示）、ID/等级/经验条
 * - 每日修行：任务三态（进行中/可领取/已领取），领取后刷新任务与资产
 * - 活动 banner 入口
 */
import { computed, onMounted, ref } from 'vue'

import { claimTask as apiClaimTask, fetchDailyTasks } from '@/api/task'
import JadeProgressBar from '@/components/ui/JadeProgressBar.vue'
import LoadingState from '@/components/ui/LoadingState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import { useLocalAvatar } from '@/composables/useLocalAvatar'
import { useToast } from '@/composables/useToast'
import { IMAGE_ACTIVITY_BANNER, IMAGE_HOME_BACKGROUND } from '@/constants/heroAssets'
import { useAssetStore } from '@/stores/assets'
import { usePlayerStore } from '@/stores/player'
import { errorText } from '@/types/api'
import { TASK_STATUS, type TaskView } from '@/types/task'

const toast = useToast()
const playerStore = usePlayerStore()
const assetStore = useAssetStore()
const { avatarSrc, handleFileChange } = useLocalAvatar((message) => toast.show(message))

const tasks = ref<TaskView[]>([])
const tasksLoading = ref(false)
const tasksError = ref<string | null>(null)
const claimingTaskId = ref<number | null>(null)

// 行内改名状态
const editingName = ref(false)
const nameInput = ref('')
const nameError = ref('')
const nameSaving = ref(false)

const profile = computed(() => playerStore.profile)

async function loadTasks(): Promise<void> {
  tasksLoading.value = true
  tasksError.value = null
  try {
    tasks.value = await fetchDailyTasks()
  } catch (error) {
    tasksError.value = errorText(error, '任务加载失败')
  } finally {
    tasksLoading.value = false
  }
}

function initHome(): void {
  void playerStore.refresh().catch(() => toast.show('资料加载失败，请刷新重试'))
  void assetStore.refresh().catch(() => undefined)
  void loadTasks()
}

onMounted(initHome)

function enterNameEdit(): void {
  editingName.value = true
  nameInput.value = playerStore.nickname
  nameError.value = ''
}

function exitNameEdit(): void {
  editingName.value = false
  nameInput.value = playerStore.nickname
  nameError.value = ''
  nameSaving.value = false
}

async function saveProfileName(): Promise<void> {
  if (nameSaving.value) return
  nameSaving.value = true
  nameError.value = ''
  try {
    await playerStore.updateName(nameInput.value)
    exitNameEdit()
  } catch (error) {
    // 服务端中文文案（长度/敏感词/每日 3 次上限等）行内展示
    nameError.value = errorText(error, '修改失败，请稍后重试')
    nameSaving.value = false
  }
}

function taskIcon(task: TaskView): string {
  return task.event_type === 'gacha_draw' ? 'auto_awesome' : 'swords'
}

/** 未完成任务的"前往"跳转：抽卡任务去召唤，其余去关卡 */
function taskTargetRoute(task: TaskView): { name: string } {
  return task.event_type === 'gacha_draw' ? { name: 'gacha' } : { name: 'stages' }
}

function taskProgressPercent(task: TaskView): number {
  return Math.min(100, Math.round(((task.progress || 0) / Math.max(1, task.target_count || 1)) * 100))
}

async function claimTask(task: TaskView): Promise<void> {
  if (claimingTaskId.value !== null) return
  claimingTaskId.value = task.task_id
  try {
    await apiClaimTask(task.task_id)
    // 与旧版 initHome() 等价：领取后整体刷新
    initHome()
  } catch (error) {
    toast.show(`领取失败：${errorText(error)}`)
  } finally {
    claimingTaskId.value = null
  }
}
</script>

<template>
  <div class="relative">
    <!-- 沉浸式背景 -->
    <div class="pointer-events-none absolute -inset-x-margin-mobile -top-stack-lg bottom-0 -z-10 overflow-hidden md:-inset-x-margin-desktop">
      <img :src="IMAGE_HOME_BACKGROUND" alt="" class="h-full w-full object-cover object-center opacity-95" />
      <div
        class="absolute inset-0 bg-[radial-gradient(circle_at_32%_18%,rgba(255,225,109,.22),transparent_24%),linear-gradient(135deg,rgba(45,45,45,.18)_0%,rgba(19,19,19,.45)_65%,rgba(14,14,14,.62)_100%)]"
      ></div>
      <div class="ink-overlay absolute inset-0"></div>
    </div>

    <div class="mx-auto max-w-7xl space-y-stack-lg">
      <!-- 玩家资料区 -->
      <section class="grid grid-cols-1 items-end gap-stack-lg pt-4 md:grid-cols-12">
        <div class="flex items-center gap-stack-lg md:col-span-7">
          <!-- 头像框 + 上传 -->
          <div class="group relative shrink-0">
            <div
              class="h-24 w-24 overflow-hidden rounded-full border-4 border-primary-container bg-ink-wash p-1 shadow-[0_0_20px_rgba(255,215,0,0.4)] md:h-32 md:w-32"
            >
              <img :src="avatarSrc" alt="取经人头像" class="h-full w-full rounded-full bg-surface-container object-cover" />
            </div>
            <label
              class="absolute -right-2 top-1 flex h-9 w-9 cursor-pointer items-center justify-center rounded-full border border-primary-container bg-surface-container-high text-primary-fixed shadow-lg transition hover:bg-primary-container hover:text-on-primary-container"
              title="更换头像"
            >
              <span class="material-symbols-outlined text-lg" aria-hidden="true">photo_camera</span>
              <input accept="image/*" class="hidden" type="file" @change="handleFileChange" />
            </label>
            <div
              class="absolute -bottom-3 left-1/2 flex h-7 min-w-[76px] -translate-x-1/2 items-center justify-center whitespace-nowrap rounded-full border border-on-primary-container/30 bg-primary-container px-3 text-center font-stats-num text-label-sm leading-none text-on-primary-container"
            >
              等级 {{ playerStore.level }}
            </div>
          </div>

          <!-- 资料明细 -->
          <div class="min-w-0 space-y-stack-xs">
            <h2 class="font-display-hero text-headline-lg-mobile text-primary drop-shadow-lg md:text-headline-lg">
              {{ playerStore.nickname }}
            </h2>
            <div class="flex flex-wrap items-center gap-2">
              <div
                class="inline-flex items-center gap-1.5 rounded border border-outline-variant bg-surface-container-low/80 px-2.5 py-1 font-label-sm text-label-sm text-on-surface-variant"
              >
                <span class="material-symbols-outlined text-sm text-primary-fixed" aria-hidden="true">badge</span>
                <span>ID</span>
                <span class="font-stats-num text-primary-fixed">{{ profile?.unique_id ?? '--' }}</span>
              </div>
              <div
                class="inline-flex items-center gap-1.5 rounded border border-outline-variant bg-surface-container-low/80 px-2.5 py-1 font-label-sm text-label-sm text-on-surface-variant"
              >
                <span class="material-symbols-outlined text-sm text-secondary-container" aria-hidden="true">swords</span>
                <span>战力</span>
                <span class="font-stats-num text-primary-fixed">{{ playerStore.power.toLocaleString('zh-CN') }}</span>
              </div>
              <button
                v-if="!editingName"
                class="inline-flex h-8 items-center gap-1.5 rounded border border-primary-container/60 bg-surface-container-high px-2.5 font-label-sm text-label-sm text-primary-fixed transition hover:bg-primary-container hover:text-on-primary-container"
                title="修改用户名"
                type="button"
                @click="enterNameEdit"
              >
                <span class="material-symbols-outlined text-sm" aria-hidden="true">edit</span>
                <span>修改用户名</span>
              </button>
            </div>

            <!-- 行内改名面板 -->
            <div v-if="editingName" class="w-full max-w-xs space-y-2">
              <div class="flex items-center gap-2">
                <input
                  v-model="nameInput"
                  :disabled="nameSaving"
                  maxlength="16"
                  type="text"
                  class="h-10 min-w-0 flex-1 rounded border border-outline-variant bg-surface-container-low px-3 font-body-md text-sm text-on-surface outline-none transition focus:border-primary-container disabled:opacity-60"
                  @keydown.enter.prevent="saveProfileName"
                  @keydown.escape.prevent="exitNameEdit"
                />
                <button
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded border border-primary-container bg-primary-container text-on-primary-container transition hover:brightness-110 disabled:opacity-60"
                  :disabled="nameSaving"
                  title="保存"
                  type="button"
                  @click="saveProfileName"
                >
                  <span class="material-symbols-outlined text-base" aria-hidden="true">{{
                    nameSaving ? 'progress_activity' : 'check'
                  }}</span>
                </button>
                <button
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded border border-outline-variant bg-surface-container-high text-on-surface-variant transition hover:text-on-surface disabled:opacity-60"
                  :disabled="nameSaving"
                  title="取消"
                  type="button"
                  @click="exitNameEdit"
                >
                  <span class="material-symbols-outlined text-base" aria-hidden="true">close</span>
                </button>
              </div>
              <p v-if="nameError" class="font-label-sm text-label-sm text-error">{{ nameError }}</p>
            </div>

            <!-- 经验条 -->
            <div class="w-full max-w-xs space-y-1 pt-1">
              <div class="flex items-center justify-between font-label-sm text-label-sm text-on-surface-variant">
                <span>经验进度</span>
                <span class="font-stats-num">{{ profile?.exp ?? 0 }} / {{ playerStore.expMax }}</span>
              </div>
              <JadeProgressBar :value="profile?.exp ?? 0" :max="playerStore.expMax" height-class="h-3" />
            </div>
          </div>
        </div>
      </section>

      <!-- 每日修行 + 活动入口 -->
      <div class="grid grid-cols-1 gap-stack-lg lg:grid-cols-3">
        <section class="space-y-stack-md lg:col-span-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-primary-fixed" aria-hidden="true">history_edu</span>
              <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-primary">每日修行</h3>
            </div>
            <span class="rounded-full bg-surface-container px-3 py-1 font-label-sm text-label-sm text-on-surface-variant"
              >每日 0 点刷新</span
            >
          </div>

          <LoadingState v-if="tasksLoading && tasks.length === 0" text="任务加载中…" />
          <ErrorState v-else-if="tasksError" :message="tasksError" @retry="loadTasks" />

          <div v-else class="space-y-stack-md">
            <div
              v-for="task in tasks"
              :key="task.task_id"
              class="flex min-h-[112px] items-center justify-between rounded-xl p-stack-lg backdrop-blur-md transition-colors"
              :class="
                task.status === TASK_STATUS.claimed
                  ? 'border border-outline-variant/30 bg-surface-container-lowest/50 opacity-60'
                  : task.status === TASK_STATUS.claimable
                    ? 'border-2 border-primary-container bg-ink-wash/90'
                    : 'border border-outline-variant bg-ink-wash/80 hover:border-primary-fixed/50'
              "
            >
              <div class="flex items-center gap-4">
                <div
                  class="flex h-12 w-12 items-center justify-center rounded border border-outline-variant bg-surface-container-high"
                >
                  <span class="material-symbols-outlined text-primary-container" aria-hidden="true">{{
                    taskIcon(task)
                  }}</span>
                </div>
                <div>
                  <p
                    class="font-title-md text-title-md"
                    :class="task.status !== TASK_STATUS.inProgress ? 'text-primary-container' : 'text-on-surface'"
                  >
                    {{ task.name }}
                  </p>
                  <p class="font-label-sm text-label-sm text-on-surface-variant">
                    进度 {{ task.progress || 0 }} / {{ task.target_count || 0 }}
                  </p>
                  <div class="mt-2 flex items-center gap-2">
                    <div class="w-32">
                      <JadeProgressBar
                        :value="task.progress || 0"
                        :max="Math.max(1, task.target_count || 1)"
                        variant="gold"
                        height-class="h-1.5"
                      />
                    </div>
                    <span class="font-stats-num text-[10px]">
                      {{
                        task.status === TASK_STATUS.claimed
                          ? '已领取'
                          : task.status === TASK_STATUS.claimable
                            ? '可领取'
                            : `${taskProgressPercent(task)}%`
                      }}
                    </span>
                  </div>
                </div>
              </div>

              <div class="flex flex-col items-end gap-2">
                <div class="flex items-center gap-3 font-stats-num text-label-sm">
                  <span v-if="task.reward_gold" class="text-primary-fixed">金币 x{{ task.reward_gold }}</span>
                  <span v-if="task.reward_diamond" class="text-quality-ssr">灵玉 x{{ task.reward_diamond }}</span>
                </div>

                <button
                  v-if="task.status === TASK_STATUS.claimed"
                  disabled
                  class="cursor-not-allowed rounded border border-outline-variant/30 bg-surface-variant px-6 py-2 font-title-md text-label-sm text-on-surface-variant/50"
                >
                  已领取
                </button>
                <button
                  v-else-if="task.status === TASK_STATUS.claimable"
                  class="jade-shimmer rounded border border-white/20 bg-gradient-to-b from-[#00A86B] to-[#007a4e] px-6 py-2 font-title-md text-label-sm text-white shadow-[0_4px_10px_rgba(0,168,107,0.3)] transition-all hover:brightness-110 active:scale-95 disabled:opacity-70"
                  :disabled="claimingTaskId !== null"
                  @click="claimTask(task)"
                >
                  <span
                    v-if="claimingTaskId === task.task_id"
                    class="material-symbols-outlined animate-spin text-sm"
                    aria-hidden="true"
                    >progress_activity</span
                  >
                  <span v-else>领取</span>
                </button>
                <router-link
                  v-else
                  :to="taskTargetRoute(task)"
                  class="border-2 border-primary-fixed bg-ink-wash px-6 py-2 font-title-md text-label-sm text-primary-fixed transition-all hover:bg-primary-fixed hover:text-on-primary-fixed active:scale-95"
                >
                  前往
                </router-link>
              </div>
            </div>
          </div>
        </section>

        <!-- 活动入口 -->
        <aside class="space-y-stack-lg">
          <router-link
            :to="{ name: 'activity' }"
            class="relative block min-h-[360px] cursor-pointer overflow-hidden rounded-xl border-2 border-outline-variant bg-surface-container transition-colors hover:border-primary-fixed/60 lg:min-h-[460px]"
          >
            <img
              :src="IMAGE_ACTIVITY_BANNER"
              alt="火焰山活动"
              class="absolute inset-0 h-full w-full object-cover object-center"
            />
            <div class="absolute inset-0 bg-gradient-to-t from-ink-wash via-ink-wash/35 to-transparent"></div>
            <div class="absolute bottom-3 left-3">
              <span class="mb-1 inline-block rounded-full bg-quality-ur px-2 py-0.5 text-[10px] font-bold text-white"
                >热门活动</span
              >
              <p class="font-title-md text-title-md leading-tight text-primary">火焰山突袭</p>
              <p class="font-label-sm text-label-sm text-secondary">奖励翻倍开启！</p>
            </div>
          </router-link>
        </aside>
      </div>
    </div>
  </div>
</template>
