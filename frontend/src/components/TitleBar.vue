<script setup lang="ts">
import SvgIcon from './SvgIcon.vue'
import { t } from '../i18n'
import { state, toast } from '../store'
import * as api from '../api'

function onWin(action: 'min' | 'max' | 'close') {
  if (!api.wailsAvailable) {
    toast(t('toast.window'), 'warning')
    return
  }
  if (action === 'min') void api.minimize()
  else if (action === 'max') void api.toggleMaximize()
  else void api.quit()
}

function onTitlebarDblclick(e: MouseEvent) {
  if ((e.target as HTMLElement).closest('.titlebar-controls')) return
  if (api.wailsAvailable) void api.toggleMaximize()
}
</script>

<template>
  <header class="titlebar" @dblclick="onTitlebarDblclick">
    <div class="titlebar-left">
      <span class="app-mark" aria-hidden="true">
        <SvgIcon name="download" :width="2" />
      </span>
      <span class="app-title">{{ t('app.name') }}</span>
      <span class="version-pill">{{ state.settings.currentVersion }}</span>
    </div>
    <div class="titlebar-controls">
      <button class="win-btn" :aria-label="t('win.min')" @click="onWin('min')">
        <SvgIcon name="minimize" />
      </button>
      <button class="win-btn" :aria-label="t('win.max')" @click="onWin('max')">
        <SvgIcon name="maximize" />
      </button>
      <button class="win-btn win-close" :aria-label="t('win.close')" @click="onWin('close')">
        <SvgIcon name="x" />
      </button>
    </div>
  </header>
</template>
