<script setup lang="ts">
import { computed } from 'vue'
import { NConfigProvider, NMessageProvider, NDialogProvider, NGlobalStyle, darkTheme } from 'naive-ui'
import Shell from './components/Shell.vue'
import { naiveLocale } from './i18n'
import { isDark } from './store'
import { lightThemeOverrides, darkThemeOverrides } from './theme'

const theme = computed(() => (isDark.value ? darkTheme : null))
const overrides = computed(() => (isDark.value ? darkThemeOverrides : lightThemeOverrides))
</script>

<template>
  <NConfigProvider :theme="theme" :theme-overrides="overrides" :locale="naiveLocale">
    <NDialogProvider>
      <NMessageProvider>
        <NGlobalStyle />
        <Shell />
      </NMessageProvider>
    </NDialogProvider>
  </NConfigProvider>
</template>
