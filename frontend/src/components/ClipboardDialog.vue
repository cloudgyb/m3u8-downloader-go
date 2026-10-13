<script setup lang="ts">
import { NModal, NCheckboxGroup, NCheckbox, NInput, NButton } from 'naive-ui'
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
        <div v-for="l in pendingLinks" :key="l.url" class="clip-item">
          <NCheckbox :value="l.url" class="clip-check">
            <span class="mono clip-url">{{ l.url }}</span>
          </NCheckbox>
          <NInput
            v-model:value="l.title"
            size="small"
            class="clip-name-input"
            :placeholder="t('clip.namePh')"
          />
        </div>
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
.clip-list { max-height: 360px; overflow-y: auto; display: flex; flex-direction: column; gap: 10px; }
.clip-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  border: 1px solid var(--border-soft);
  border-radius: var(--radius-sm);
}
.clip-check { flex: 1; min-width: 0; }
.clip-check :deep(.n-checkbox__label) { white-space: normal; }
.clip-url { word-break: break-all; }
.clip-name-input { width: 200px; flex: none; }
.clip-actions { display: flex; justify-content: flex-end; gap: 10px; }
</style>
