<script setup lang="ts">
import { computed } from 'vue'
import { ICONS } from '../icons'

const props = withDefaults(defineProps<{ name: string; width?: number }>(), { width: 1.6 })

const paths = computed(() => ICONS[props.name] ?? '')
</script>

<template>
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    :stroke-width="width"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    v-html="paths"
  />
</template>
