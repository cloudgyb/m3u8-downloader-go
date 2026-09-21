<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NInput, NButton, NSelect } from 'naive-ui'
import SvgIcon from './SvgIcon.vue'
import Segmented from './Segmented.vue'
import { t } from '../i18n'
import { state, addTask, updateSettings } from '../store'
import * as api from '../api'
import type { Variant } from '../types'

const url = ref('')
const name = ref('')
const error = ref<string | null>(null)

const variants = ref<Variant[]>([])
const selectedVariant = ref<string | null>(null)
const loadingVariants = ref(false)

const threads = computed<string>({
  get: () => String(state.settings.threads),
  set: (v: string) => updateSettings({ threads: parseInt(v, 10) }, true),
})

const threadOptions = [
  { value: '4', label: '4' },
  { value: '8', label: '8' },
  { value: '16', label: '16' },
  { value: '32', label: '32' },
]

const proxySummary = computed(() =>
  state.settings.proxyEnabled
    ? `${state.settings.proxyHost}:${state.settings.proxyPort} (${state.settings.proxyType.toUpperCase()})`
    : t('status.noProxy'),
)

function fmtBandwidth(b: number): string {
  if (b >= 1e6) return (b / 1e6).toFixed(1) + ' Mbps'
  return Math.round(b / 1000) + ' Kbps'
}

function variantLabel(v: Variant): string {
  const res = v.resolution || t('dl.variant.unknown')
  return v.bandwidth > 0 ? res + ' · ' + fmtBandwidth(v.bandwidth) : res
}

const variantOptions = computed(() => variants.value.map((v) => ({ value: v.uri, label: variantLabel(v) })))

async function loadVariants() {
  const u = url.value.trim()
  if (!/^https?:\/\//i.test(u)) {
    variants.value = []
    selectedVariant.value = null
    return
  }
  loadingVariants.value = true
  try {
    const list = await api.fetchVariants(u, {
      enabled: state.settings.proxyEnabled,
      type: state.settings.proxyType,
      host: state.settings.proxyHost,
      port: state.settings.proxyPort,
    })
    variants.value = list
    if (list.length > 1) {
      const best = list.reduce((a, b) => (b.bandwidth > a.bandwidth ? b : a), list[0])
      selectedVariant.value = best.uri
    } else {
      selectedVariant.value = list[0]?.uri ?? null
    }
  } catch {
    variants.value = []
    selectedVariant.value = null
  } finally {
    loadingVariants.value = false
  }
}

watch(url, () => {
  variants.value = []
  selectedVariant.value = null
})

async function start() {
  const err = await addTask(url.value, name.value, selectedVariant.value ?? '')
  if (err) {
    error.value = t(err)
    return
  }
  error.value = null
  url.value = ''
  name.value = ''
  variants.value = []
  selectedVariant.value = null
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Enter') void start()
}
</script>

<template>
  <div class="page-head">
    <h1>{{ t('dl.title') }}</h1>
    <p class="lead">{{ t('dl.subtitle') }}</p>
  </div>
  <div class="dl-grid">
    <div class="card dl-form-card">
      <div class="field">
        <label class="label" for="dlUrl">{{ t('dl.url') }}</label>
        <NInput
          id="dlUrl"
          v-model:value="url"
          class="mono-input"
          :placeholder="'https://example.com/video/index.m3u8'"
          spellcheck="false"
          autocomplete="off"
          @blur="loadVariants"
          @keydown="onKey"
        />
        <p v-if="error" class="error">{{ error }}</p>
      </div>
      <div class="field">
        <label class="label" for="dlName">{{ t('dl.name') }}</label>
        <NInput
          id="dlName"
          v-model:value="name"
          :placeholder="t('dl.namePh')"
          spellcheck="false"
          @keydown="onKey"
        />
      </div>
      <div v-if="variants.length > 1" class="field">
        <label class="label">{{ t('dl.variant') }}</label>
        <NSelect v-model:value="selectedVariant" :options="variantOptions" :loading="loadingVariants" />
      </div>
      <div class="field">
        <label class="label">{{ t('dl.threads') }}</label>
        <Segmented v-model="threads" :options="threadOptions" />
      </div>
      <div class="field">
        <label class="label">{{ t('dl.saveTo') }}</label>
        <div class="readonly-box mono">{{ state.settings.saveDir }}</div>
      </div>
      <NButton type="primary" size="large" block @click="start">
        <template #icon><SvgIcon name="download" :width="1.8" /></template>
        {{ t('dl.start') }}
      </NButton>
    </div>

    <aside class="card dl-side">
      <h3>{{ t('dl.summaryTitle') }}</h3>
      <div class="kv"><span>{{ t('dl.threads') }}</span><b>{{ state.settings.threads }}</b></div>
      <div class="kv"><span>{{ t('status.proxy') }}</span><b>{{ proxySummary }}</b></div>
      <div class="kv"><span>{{ t('dl.saveTo') }}</span><b class="mono">{{ state.settings.saveDir }}</b></div>
      <div class="kv"><span>{{ t('set.update.current') }}</span><b class="mono">{{ state.settings.currentVersion }}</b></div>
    </aside>
  </div>
</template>
