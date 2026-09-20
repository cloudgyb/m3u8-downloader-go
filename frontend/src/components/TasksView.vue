<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NProgress, NEmpty } from 'naive-ui'
import SvgIcon from './SvgIcon.vue'
import { t } from '../i18n'
import {
  state, navigate, pauseTask, resumeTask, cancelTask, pauseAll, resumeAll,
  activeCount, totalSpeed, doneCount,
} from '../store'
import { fmtBytes, fmtSpeed, fmtEta } from '../utils'
import { iconNameForStatus, toneForStatus } from '../icons'
import type { Task } from '../types'

function pct(task: Task): number {
  if (task.total <= 0) return 0
  return Math.min(100, (task.done / task.total) * 100)
}
function speedText(task: Task): string {
  return task.status === 'downloading' ? t('row.speed') + ' ' + fmtSpeed(task.speed) : '—'
}
function etaText(task: Task): string {
  return task.status === 'downloading' && task.speed > 0
    ? t('row.eta') + ' ' + fmtEta((task.total - task.done) / task.speed)
    : '—'
}
function badgeClass(status: string): string {
  const tone = toneForStatus(status)
  if (tone === 'success') return 'badge-success'
  if (tone === 'danger') return 'badge-danger'
  if (tone === 'downloading') return 'badge-accent'
  return 'badge-muted'
}
</script>

<template>
  <div class="page-head tasks-head">
    <div>
      <h1>{{ t('tasks.title') }}</h1>
      <div class="tasks-stats">
        <div class="od-stat">
          <span class="stat-num">{{ activeCount }}</span>
          <span class="stat-label">{{ t('tasks.active') }}</span>
        </div>
        <div class="od-stat">
          <span class="stat-num">{{ activeCount > 0 ? fmtSpeed(totalSpeed) : '0 B/s' }}</span>
          <span class="stat-label">{{ t('tasks.totalSpeed') }}</span>
        </div>
        <div class="od-stat">
          <span class="stat-num">{{ doneCount }}</span>
          <span class="stat-label">{{ t('tasks.finished') }}</span>
        </div>
      </div>
    </div>
    <div class="tasks-actions">
      <NButton secondary @click="pauseAll()">
        <template #icon><SvgIcon name="pause" /></template>
        {{ t('tasks.pauseAll') }}
      </NButton>
      <NButton secondary @click="resumeAll()">
        <template #icon><SvgIcon name="play" /></template>
        {{ t('tasks.resumeAll') }}
      </NButton>
    </div>
  </div>

  <div class="card list">
    <div v-if="!state.tasks.length" class="empty">
      <SvgIcon name="download" />
      <p>{{ t('tasks.empty') }}</p>
      <p class="hint">{{ t('tasks.emptyHint') }}</p>
      <NButton type="primary" @click="navigate('download')">{{ t('tasks.go') }}</NButton>
    </div>

    <div v-for="task in state.tasks" :key="task.id" class="task-row">
      <div class="task-icon" :class="'tone-' + toneForStatus(task.status)">
        <SvgIcon :name="iconNameForStatus(task.status)" />
      </div>
      <div class="task-main">
        <div class="task-title od-truncate" :title="task.name">{{ task.name }}</div>
        <NProgress
          type="line"
          :percentage="pct(task)"
          :show-indicator="false"
          :height="8"
          :border-radius="9999"
        />
        <div class="task-meta">
          <span class="pct nowrap">{{ pct(task).toFixed(0) }}%</span>
          <span class="nowrap">{{ fmtBytes(task.done) }} / {{ fmtBytes(task.total) }}</span>
          <span class="nowrap">{{ speedText(task) }}</span>
          <span class="nowrap">{{ etaText(task) }}</span>
        </div>
      </div>
      <div class="task-badge">
        <span class="badge" :class="badgeClass(task.status)">{{ t('st.' + task.status) }}</span>
      </div>
      <div class="task-actions">
        <button
          v-if="task.status === 'downloading'"
          class="icon-btn"
          :title="t('row.pause')"
          :aria-label="t('row.pause')"
          @click="pauseTask(task.id)"
        >
          <SvgIcon name="pause" />
        </button>
        <button
          v-else-if="task.status !== 'queued'"
          class="icon-btn"
          :title="t('row.resume')"
          :aria-label="t('row.resume')"
          @click="resumeTask(task.id)"
        >
          <SvgIcon name="play" />
        </button>
        <button
          class="icon-btn danger"
          :title="t('row.cancel')"
          :aria-label="t('row.cancel')"
          @click="cancelTask(task.id)"
        >
          <SvgIcon name="x" />
        </button>
      </div>
    </div>
  </div>
</template>
