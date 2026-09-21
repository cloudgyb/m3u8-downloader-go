<script setup lang="ts">
import {computed} from 'vue'
import {NButton} from 'naive-ui'
import SvgIcon from './SvgIcon.vue'
import Segmented from './Segmented.vue'
import {t} from '../i18n'
import {
  clearHistory,
  filteredHistory,
  openFolder,
  redoTask,
  refreshHistory,
  removeHistoryItem,
  setHistFilter,
  state
} from '../store'
import {fmtBytes} from '../utils'
import {iconNameForStatus, toneForStatus} from '../icons'
import type {HistFilter, HistoryItem} from '../types'

const filter = computed<string>({
  get: () => state.histFilter,
  set: (v: string) => setHistFilter(v as HistFilter),
})

const filterOptions = [
  {value: 'all', label: t('hist.filter.all')},
  {value: 'done', label: t('hist.filter.done')},
  {value: 'fail', label: t('hist.filter.fail')},
]

function sizeText(h: HistoryItem): string {
  return h.status === 'completed' || h.size > 0 ? fmtBytes(h.size) : '—'
}

function badgeClass(status: string): string {
  return status === 'completed' ? 'badge-success' : 'badge-danger'
}
</script>

<template>
  <div class="page-head tasks-head">
    <div><h1>{{ t('hist.title') }}</h1></div>
    <div class="tasks-actions">
      <Segmented v-model="filter" :options="filterOptions"/>
      <NButton secondary @click="refreshHistory()">
        <template #icon>
          <SvgIcon name="refresh"/>
        </template>
        {{ t('toast.refresh') }}
      </NButton>
      <NButton secondary @click="clearHistory()">
        <template #icon>
          <SvgIcon name="trash"/>
        </template>
        {{ t('hist.clear') }}
      </NButton>
    </div>
  </div>

  <div class="card list">
    <div v-if="!filteredHistory.length" class="empty">
      <SvgIcon name="clock"/>
      <p>{{ t('hist.empty') }}</p>
    </div>

    <div v-for="h in filteredHistory" :key="h.id" class="hist-row">
      <div class="task-icon" :class="'tone-' + toneForStatus(h.status)">
        <SvgIcon :name="iconNameForStatus(h.status)"/>
      </div>
      <div class="task-main">
        <div class="task-title od-truncate" :title="h.name">{{ h.name }}</div>
        <div class="task-meta">
          <span class="nowrap">{{ sizeText(h) }}</span>
          <span class="nowrap">{{ t('hist.time') }} {{ h.finishedAt }}</span>
          <span class="nowrap">{{ t('hist.duration') }} {{ h.duration }}</span>
          <span v-if="h.error" class="err nowrap">{{ t(h.error) }}</span>
        </div>
      </div>
      <div class="task-badge">
        <span class="badge" :class="badgeClass(h.status)">{{ t('st.' + h.status) }}</span>
      </div>
      <div class="task-actions">
        <button
            v-if="h.status !== 'completed'"
            class="icon-btn"
            :title="t('hist.redownload')"
            :aria-label="t('hist.redownload')"
            @click="redoTask(h)"
        >
          <SvgIcon name="refresh"/>
        </button>
        <button
            class="icon-btn"
            :title="t('hist.openFolder')"
            :aria-label="t('hist.openFolder')"
            @click="openFolder(h)"
        >
          <SvgIcon name="folder"/>
        </button>

        <button
            class="icon-btn"
            :title="t('toast.delete')"
            :aria-label="t('toast.delete')"
            @click="removeHistoryItem(h.id)"
        >
          <SvgIcon name="trash"/>
        </button>
      </div>
    </div>
  </div>
</template>
