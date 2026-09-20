<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from './SvgIcon.vue'
import { t } from '../i18n'
import { state, activeCount, totalSpeed, statusText } from '../store'
import { fmtSpeed } from '../utils'

const proxyText = computed(() =>
  state.settings.proxyEnabled
    ? t('status.proxy') + ' ' + state.settings.proxyHost + ':' + state.settings.proxyPort
    : t('status.noProxy'),
)
const speedText = computed(() => (activeCount.value > 0 ? t('status.total') + ' ' + fmtSpeed(totalSpeed.value) : ''))
</script>

<template>
  <footer class="statusbar">
    <div class="sb-group">
      <span class="sb-dot" :class="{ on: activeCount > 0 }"></span>
      <span>{{ statusText }}</span>
    </div>
    <div class="sb-group">
      <span class="sb-item">
        <SvgIcon name="globe" />
        <span>{{ proxyText }}</span>
      </span>
      <span v-if="speedText" class="sb-item mono">{{ speedText }}</span>
      <span class="sb-item mono">{{ state.settings.currentVersion }}</span>
    </div>
  </footer>
</template>
