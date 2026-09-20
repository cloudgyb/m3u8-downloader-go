<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { NModal } from 'naive-ui'
import { t, lang, releaseNotes } from '../i18n'
import { updateDialogVisible, toast } from '../store'

const phase = ref<'checking' | 'done'>('checking')
const notes = computed(() => releaseNotes[lang.value])

watch(updateDialogVisible, (v) => {
  if (v) {
    phase.value = 'checking'
    setTimeout(() => {
      if (updateDialogVisible.value) phase.value = 'done'
    }, 1100)
  }
})

function install() {
  updateDialogVisible.value = false
  toast(t('toast.update'), 'info')
}
</script>

<template>
  <NModal
    v-model:show="updateDialogVisible"
    preset="dialog"
    :title="t('set.update.check')"
    :positive-text="t('upd.install')"
    :negative-text="t('upd.later')"
    @positive-click="install"
    @negative-click="updateDialogVisible = false"
  >
    <div v-if="phase === 'checking'" class="upd-checking">
      <span class="spinner"></span>
      <span>{{ t('upd.checking') }}</span>
    </div>
    <div v-else class="upd-result">
      <span class="upd-ver">v1.3.0</span>
      <h3>{{ t('upd.found') }}</h3>
      <ul class="upd-notes">
        <li v-for="n in notes" :key="n">{{ n }}</li>
      </ul>
    </div>
  </NModal>
</template>

<style scoped>
.upd-checking { display: flex; align-items: center; gap: 12px; color: var(--fg-2); font-size: 14px; padding: 6px 0; }
.spinner { width: 18px; height: 18px; border-radius: 50%; border: 2px solid var(--border); border-top-color: var(--accent); animation: spin 0.8s linear infinite; flex: none; }
@keyframes spin { to { transform: rotate(360deg); } }
.upd-ver { font-family: var(--font-mono); font-size: 13px; color: var(--muted); }
.upd-result h3 { margin: 6px 0 12px; font-size: 17px; font-weight: 700; }
.upd-notes { margin: 0; padding-left: 18px; font-size: 13px; color: var(--fg-2); display: flex; flex-direction: column; gap: 6px; line-height: 1.5; }
</style>
