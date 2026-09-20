<script lang="ts">
// Module-scoped counter: every instance gets a unique radio-group name so
// multiple segmented controls on the page don't interfere with each other.
let segSeq = 0
</script>

<script setup lang="ts">
const name = 'seg-' + ++segSeq

defineProps<{ modelValue: string; options: { value: string; label: string }[] }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()
</script>

<template>
  <div class="segmented" role="radiogroup">
    <label v-for="o in options" :key="o.value" class="seg">
      <input
        type="radio"
        :name="name"
        :value="o.value"
        :checked="o.value === modelValue"
        @change="emit('update:modelValue', o.value)"
      />
      <span>{{ o.label }}</span>
    </label>
  </div>
</template>
