<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from './SvgIcon.vue'
import { t } from '../i18n'

const props = withDefaults(defineProps<{ page: number; pageCount: number; total?: number }>(), {
  total: 0,
})
const emit = defineEmits<{ (e: 'update:page', v: number): void }>()

function go(p: number) {
  if (p < 1 || p > props.pageCount || p === props.page) return
  emit('update:page', p)
}

// The ellipsis marker is only ever rendered as a span, but TS can't narrow it
// in the template, so pageNum resolves the union for the button click.
function pageNum(p: number | '…'): number {
  return typeof p === 'number' ? p : props.page
}

// Page numbers with a single ellipsis gap on each side when there are many pages.
const pages = computed<(number | '…')[]>(() => {
  const count = props.pageCount
  if (count <= 7) return Array.from({ length: count }, (_, i) => i + 1)
  const wanted = new Set<number>([1, count, props.page - 1, props.page, props.page + 1])
  const nums = [...wanted].filter((n) => n >= 1 && n <= count).sort((a, b) => a - b)
  const out: (number | '…')[] = []
  let prev = 0
  for (const n of nums) {
    if (n - prev > 1) out.push('…')
    out.push(n)
    prev = n
  }
  return out
})
</script>

<template>
  <div class="pagination">
    <span v-if="total > 0" class="pg-total">{{ t('pg.total') }} {{ total }}</span>
    <template v-if="pageCount > 1">
      <button
        class="pg-btn"
        :disabled="page <= 1"
        :aria-label="t('pg.prev')"
        :title="t('pg.prev')"
        @click="go(page - 1)"
      >
        <SvgIcon name="chevron-left" />
      </button>
      <template v-for="(p, i) in pages" :key="i">
        <span v-if="p === '…'" class="pg-ellipsis">…</span>
        <button v-else class="pg-btn" :class="{ on: p === page }" @click="go(pageNum(p))">{{ p }}</button>
      </template>
      <button
        class="pg-btn"
        :disabled="page >= pageCount"
        :aria-label="t('pg.next')"
        :title="t('pg.next')"
        @click="go(page + 1)"
      >
        <SvgIcon name="chevron-right" />
      </button>
    </template>
  </div>
</template>
