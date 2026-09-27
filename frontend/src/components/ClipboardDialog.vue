<script setup lang="ts">
import { NModal, NCheckboxGroup, NCheckbox, NButton } from 'naive-ui'
import { t } from '../i18n'
import {
  pendingLinks,
  clipDialogVisible,
  checkedLinks,
  confirmClipboardDownload,
  cancelClipboardDownload,
} from '../store'

function onShow(v: boolean) {
  if (!v) cancelClipboardDownload()
}
</script>

<template>
  <NModal
    :show="clipDialogVisible"
    preset="card"
    class="clip-dialog"
    :title="t('clip.title')"
    style="width: 560px; max-width: calc(100vw - 48px)"
    @update:show="onShow"
  >
    <p class="clip-hint">{{ t('clip.hint') }}</p>
    <div class="clip-list">
      <NCheckboxGroup v-model:value="checkedLinks">
        <NCheckbox v-for="u in pendingLinks" :key="u" :value="u" class="clip-item">
          <span class="mono clip-url">{{ u }}</span>
        </NCheckbox>
      </NCheckboxGroup>
    </div>
    <template #footer>
      <div class="clip-actions">
        <NButton @click="cancelClipboardDownload">{{ t('clip.cancel') }}</NButton>
        <NButton type="primary" :disabled="checkedLinks.length === 0" @click="confirmClipboardDownload">
          {{ t('clip.download') }} ({{ checkedLinks.length }})
        </NButton>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
.clip-hint { font-size: 13px; color: var(--fg-2); margin: 0 0 14px; }
.clip-list { max-height: 320px; overflow-y: auto; display: flex; flex-direction: column; gap: 4px; }
.clip-item { width: 100%; }
.clip-url { word-break: break-all; }
.clip-actions { display: flex; justify-content: flex-end; gap: 10px; }
</style>
