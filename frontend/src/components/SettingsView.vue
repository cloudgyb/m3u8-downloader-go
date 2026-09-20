<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { NInput, NInputNumber, NSelect, NSwitch, NButton } from 'naive-ui'
import Segmented from './Segmented.vue'
import SvgIcon from './SvgIcon.vue'
import { t } from '../i18n'
import { state, updateSettings, setTheme, setLang, showUpdateDialog, toast, promoteQueued } from '../store'
import * as api from '../api'
import { clampInt } from '../utils'
import type { Theme, Lang } from '../types'

const threads = computed<number>({
  get: () => state.settings.threads,
  set: (v) => { if (v != null) updateSettings({ threads: clampInt(v, 1, 32) }) },
})
const concurrent = computed<number>({
  get: () => state.settings.concurrent,
  set: (v) => { if (v != null) { updateSettings({ concurrent: clampInt(v, 1, 8) }); promoteQueued() } },
})
const saveDir = computed<string>({
  get: () => state.settings.saveDir,
  set: (v) => updateSettings({ saveDir: v }),
})
const notify = computed<boolean>({
  get: () => state.settings.notify,
  set: (v) => updateSettings({ notify: v }, true),
})
const format = computed<'ts' | 'mp4'>({
  get: () => state.settings.format,
  set: (v) => updateSettings({ format: v }, true),
})
const ffmpegOk = ref(true)
onMounted(async () => {
  ffmpegOk.value = await api.hasFFmpeg()
})
const proxyEnabled = computed<boolean>({
  get: () => state.settings.proxyEnabled,
  set: (v) => updateSettings({ proxyEnabled: v }, true),
})
const proxyPort = computed<number>({
  get: () => state.settings.proxyPort,
  set: (v) => { if (v != null) updateSettings({ proxyPort: clampInt(v, 1, 65535) }) },
})
const proxyType = computed<string>({
  get: () => state.settings.proxyType,
  set: (v) => updateSettings({ proxyType: v as 'http' | 'socks5' }, true),
})
const proxyHost = computed<string>({
  get: () => state.settings.proxyHost,
  set: (v) => updateSettings({ proxyHost: v }),
})

const theme = computed<string>({
  get: () => state.theme,
  set: (v) => setTheme(v as Theme),
})
const updateMode = computed<string>({
  get: () => state.settings.updateMode,
  set: (v) => updateSettings({ updateMode: v as 'auto' | 'manual' }, true),
})

const themeOptions = [
  { value: 'light', label: t('theme.light') },
  { value: 'dark', label: t('theme.dark') },
  { value: 'system', label: t('theme.system') },
]
const updateModeOptions = [
  { value: 'auto', label: t('upd.auto') },
  { value: 'manual', label: t('upd.manual') },
]
const proxyTypeOptions = [
  { value: 'http', label: 'HTTP' },
  { value: 'socks5', label: 'SOCKS5' },
]
const langOptions = [
  { value: 'zh', label: '中文' },
  { value: 'en', label: 'English' },
]
const formatOptions = [
  { value: 'mp4', label: t('set.format.mp4') },
  { value: 'ts', label: t('set.format.ts') },
]

async function browse() {
  if (api.wailsAvailable) {
    const dir = await api.browseDir()
    if (dir) updateSettings({ saveDir: dir }, true)
  } else {
    toast(t('toast.browse'), 'warning')
  }
}
</script>

<template>
  <div class="page-head"><h1>{{ t('set.title') }}</h1></div>
  <div class="settings-grid">
    <div class="card">
      <h3>{{ t('set.sec.download') }}</h3>
      <div class="field">
        <label class="label">{{ t('set.threads') }}</label>
        <NInputNumber v-model:value="threads" :min="1" :max="32" style="width: 100%" />
        <p class="hint">{{ t('set.threadsHint') }}</p>
      </div>
      <div class="field">
        <label class="label">{{ t('set.concurrent') }}</label>
        <NInputNumber v-model:value="concurrent" :min="1" :max="8" style="width: 100%" />
      </div>
      <div class="field">
        <label class="label">{{ t('set.saveDir') }}</label>
        <div class="input-row">
          <NInput v-model:value="saveDir" class="mono-input" spellcheck="false" />
          <NButton secondary @click="browse">{{ t('set.browse') }}</NButton>
        </div>
      </div>
      <div class="field switch-row">
        <div><label class="label">{{ t('set.notify') }}</label></div>
        <NSwitch v-model:value="notify" />
      </div>
      <div class="field">
        <label class="label">{{ t('set.format') }}</label>
        <NSelect v-model:value="format" :options="formatOptions" />
        <p v-if="state.settings.format === 'mp4' && !ffmpegOk" class="hint">{{ t('set.format.ffmpegHint') }}</p>
      </div>
    </div>

    <div class="card">
      <h3>{{ t('set.sec.network') }}</h3>
      <div class="field switch-row">
        <div><label class="label">{{ t('set.proxy.enable') }}</label></div>
        <NSwitch v-model:value="proxyEnabled" />
      </div>
      <div class="field-row">
        <div class="field">
          <label class="label">{{ t('set.proxy.type') }}</label>
          <NSelect v-model:value="proxyType" :options="proxyTypeOptions" :disabled="!state.settings.proxyEnabled" />
        </div>
        <div class="field">
          <label class="label">{{ t('set.proxy.port') }}</label>
          <NInputNumber v-model:value="proxyPort" :min="1" :max="65535" :disabled="!state.settings.proxyEnabled" style="width: 100%" />
        </div>
      </div>
      <div class="field">
        <label class="label">{{ t('set.proxy.host') }}</label>
        <NInput v-model:value="proxyHost" class="mono-input" placeholder="127.0.0.1" :disabled="!state.settings.proxyEnabled" spellcheck="false" />
      </div>
    </div>

    <div class="card">
      <h3>{{ t('set.sec.appearance') }}</h3>
      <div class="field">
        <label class="label">{{ t('set.theme') }}</label>
        <Segmented v-model="theme" :options="themeOptions" />
      </div>
      <div class="field">
        <label class="label">{{ t('set.lang') }}</label>
        <NSelect :value="state.lang" :options="langOptions" @update:value="(v) => setLang(v as Lang)" />
      </div>
    </div>

    <div class="card">
      <h3>{{ t('set.sec.update') }}</h3>
      <div class="field">
        <label class="label">{{ t('set.update.mode') }}</label>
        <Segmented v-model="updateMode" :options="updateModeOptions" />
      </div>
      <div class="field switch-row">
        <div><label class="label">{{ t('set.update.current') }}</label></div>
        <span class="chip">{{ state.settings.currentVersion }}</span>
      </div>
      <NButton type="primary" @click="showUpdateDialog()">
        <template #icon><SvgIcon name="up" /></template>
        {{ t('set.update.check') }}
      </NButton>
    </div>
  </div>
</template>
