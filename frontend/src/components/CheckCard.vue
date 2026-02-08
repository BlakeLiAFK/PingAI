<script setup lang="ts">
import { ref, computed } from 'vue'
import type { FullCheckResult } from '../types'
import { t, checkItemName } from '../i18n'

const props = defineProps<{
  result: FullCheckResult
}>()

const PREVIEW_COUNT = 20
const expanded = ref(false)

const sortedModels = computed(() => {
  if (!props.result.modelList) return []
  return [...props.result.modelList].sort((a, b) =>
    a.localeCompare(b, undefined, { sensitivity: 'base' })
  )
})

const displayModels = computed(() => {
  if (expanded.value) return sortedModels.value
  return sortedModels.value.slice(0, PREVIEW_COUNT)
})

const hasMore = computed(() => sortedModels.value.length > PREVIEW_COUNT)
const moreCount = computed(() => sortedModels.value.length - PREVIEW_COUNT)

function formatLatency(ms: number): string {
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}
</script>

<template>
  <div class="check-card">
    <div class="check-card-header">
      <div class="check-card-title">
        <h3>{{ result.providerName }}</h3>
        <span class="model-tag">{{ result.model }}</span>
      </div>
      <div class="check-card-meta">
        <span>{{ t('card.totalLatency') }} {{ formatLatency(result.totalLatency) }}</span>
      </div>
    </div>

    <div class="check-items">
      <div
        v-for="item in result.results"
        :key="item.item"
        class="check-item"
        :class="item.status"
        :title="item.detail"
      >
        <span class="status-dot" :class="item.status"></span>
        <div class="item-info">
          <div class="item-name">{{ checkItemName(item.item) }}</div>
          <div class="item-msg">{{ item.message }}</div>
        </div>
        <div class="item-latency" v-if="item.latency > 0">
          {{ formatLatency(item.latency) }}
        </div>
      </div>
    </div>

    <div class="model-list-section" v-if="sortedModels.length > 0">
      <h4>{{ t('card.availableModels') }} ({{ sortedModels.length }})</h4>
      <div class="model-tags">
        <span v-for="m in displayModels" :key="m">{{ m }}</span>
        <span
          v-if="hasMore && !expanded"
          class="model-more"
          @click="expanded = true"
        >+{{ moreCount }} more</span>
        <span
          v-if="hasMore && expanded"
          class="model-more"
          @click="expanded = false"
        >{{ t('card.collapse') }}</span>
      </div>
    </div>
  </div>
</template>
