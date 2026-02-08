<script setup lang="ts">
import { ref, computed } from 'vue'
import { isBatchRunning, batchKeyResults, runBatchKeyCheck } from '../stores/check'
import { t, checkItemName } from '../i18n'
import type { FullCheckResult } from '../types'

const emit = defineEmits<{ (e: 'close'): void }>()

const keysText = ref('')
const expandedIdx = ref<number | null>(null)

const keyCount = computed(() => {
  return parseKeys(keysText.value).length
})

function parseKeys(text: string): string[] {
  return text
    .split('\n')
    .map(l => l.trim())
    .filter(l => l.length > 0)
}

async function handleRun() {
  const keys = parseKeys(keysText.value)
  if (keys.length === 0) return
  await runBatchKeyCheck(keys)
}

function overallStatus(r: FullCheckResult): string {
  const hasError = r.results.some(i => i.status === 'failed')
  if (hasError) return 'failed'
  const hasWarn = r.results.some(i => i.status === 'warning')
  if (hasWarn) return 'warning'
  return 'success'
}

function statusLabel(s: string): string {
  if (s === 'success') return 'OK'
  if (s === 'failed') return 'FAIL'
  return 'WARN'
}

function fmtLatency(ms: number): string {
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}

function toggleExpand(idx: number) {
  expandedIdx.value = expandedIdx.value === idx ? null : idx
}
</script>

<template>
  <div class="dialog-overlay" @click.self="emit('close')">
    <div class="dialog batch-dialog">
      <div class="dialog-header">
        <h3>{{ t('batchKey.title') }}</h3>
        <button class="btn-icon-sm" @click="emit('close')">&times;</button>
      </div>

      <div class="dialog-body">
        <div class="form-group">
          <label>{{ t('batchKey.label') }}</label>
          <textarea
            v-model="keysText"
            rows="6"
            placeholder="sk-key1&#10;sk-key2&#10;sk-key3"
            :disabled="isBatchRunning"
          ></textarea>
        </div>
        <div class="batch-info">
          <span>{{ t('batchKey.count', { n: keyCount }) }}</span>
          <button
            class="btn btn-primary"
            :disabled="isBatchRunning || keyCount === 0"
            @click="handleRun"
          >
            <span v-if="isBatchRunning" class="spinner"></span>
            <span v-else>{{ t('batchKey.start') }}</span>
          </button>
        </div>
      </div>

      <!-- 结果区域 -->
      <div class="batch-results" v-if="batchKeyResults.length > 0">
        <div class="batch-results-header">{{ t('batchKey.results') }}</div>
        <div class="batch-results-list">
          <div
            v-for="(r, idx) in batchKeyResults"
            :key="idx"
            class="batch-result-row"
          >
            <div class="batch-result-main" @click="toggleExpand(idx)">
              <span class="history-status" :class="overallStatus(r)">
                {{ statusLabel(overallStatus(r)) }}
              </span>
              <span class="batch-key-name">{{ r.providerName }}</span>
              <span class="history-latency">{{ fmtLatency(r.totalLatency) }}</span>
            </div>
            <div v-if="expandedIdx === idx" class="batch-result-detail">
              <div class="check-items">
                <div
                  v-for="item in r.results"
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
                  <div class="item-latency" v-if="item.latency > 0">{{ fmtLatency(item.latency) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
