<script setup lang="ts">
import { onMounted } from 'vue'
import { useMessage, useDialog } from 'naive-ui'
import TitleBar from './TitleBar.vue'
import SideBar from './SideBar.vue'
import StatusBar from './StatusBar.vue'
import DownloadView from './DownloadView.vue'
import TasksView from './TasksView.vue'
import HistoryView from './HistoryView.vue'
import SettingsView from './SettingsView.vue'
import UpdateDialog from './UpdateDialog.vue'
import { ui, state, init } from '../store'

// Expose Naive UI message/dialog to the store so non-component code can toast.
ui.message = useMessage()
ui.dialog = useDialog()

onMounted(() => {
  void init()
})
</script>

<template>
  <div class="window">
    <TitleBar />
    <div class="app">
      <SideBar />
      <main class="content">
        <section v-show="state.currentView === 'download'" class="view"><DownloadView /></section>
        <section v-show="state.currentView === 'tasks'" class="view"><TasksView /></section>
        <section v-show="state.currentView === 'history'" class="view"><HistoryView /></section>
        <section v-show="state.currentView === 'settings'" class="view"><SettingsView /></section>
      </main>
    </div>
    <StatusBar />
    <UpdateDialog />
  </div>
</template>
