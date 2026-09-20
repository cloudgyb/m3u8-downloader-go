<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from './SvgIcon.vue'
import { t } from '../i18n'
import { state, navigate, toggleTheme, toggleLang, resolvedTheme } from '../store'
import type { View } from '../types'

const items: { view: View; icon: string; label: string }[] = [
  { view: 'download', icon: 'download', label: 'nav.download' },
  { view: 'tasks', icon: 'list', label: 'nav.tasks' },
  { view: 'history', icon: 'clock', label: 'nav.history' },
  { view: 'settings', icon: 'settings', label: 'nav.settings' },
]

const themeIcon = computed(() => (resolvedTheme() === 'dark' ? 'moon' : 'sun'))
const langLabel = computed(() => (state.lang === 'zh' ? 'EN' : '中文'))
</script>

<template>
  <aside class="sidebar">
    <nav class="sidebar-nav" aria-label="main">
      <button
        v-for="item in items"
        :key="item.view"
        class="nav-item"
        :class="{ active: state.currentView === item.view }"
        @click="navigate(item.view)"
      >
        <SvgIcon :name="item.icon" />
        <span>{{ t(item.label) }}</span>
      </button>
    </nav>
    <div class="sidebar-foot">
      <button class="sb-btn" @click="toggleTheme()">
        <SvgIcon :name="themeIcon" />
        <span>{{ t('set.theme') }}</span>
      </button>
      <button class="sb-btn" @click="toggleLang()">
        <SvgIcon name="globe" />
        <span>{{ langLabel }}</span>
      </button>
    </div>
  </aside>
</template>
