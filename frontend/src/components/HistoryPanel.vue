<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { historyItems, historyTotal, loadHistory, deleteHistoryItem, deleteHistoryBatch, clearAllHistory } from '../stores/check'
import { t, checkItemName } from '../i18n'
import type { CheckResult } from '../types'

const selectedIds = ref<Set<number>>(new Set())
const expandedId = ref<number | null>(null)

onMounted(() => loadHistory())

function toggleSelect(id: number) {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
}

function selectAll() {
  if (selectedIds.value.size === historyItems.value.length) {
    selectedIds.value.clear()
  } else {
    selectedIds.value = new Set(historyItems.value.map(h => h.id))
  }
}

async function handleDeleteSelected() {
  if (selectedIds.value.size === 0) return
  if (!confirm(t('history.confirmDelete', { n: selectedIds.value.size }))) return
  await deleteHistoryBatch(Array.from(selectedIds.value))
  selectedIds.value.clear()
}

async function handleClearAll() {
  if (!confirm(t('history.confirmClear'))) return
  await clearAllHistory()
  selectedIds.value.clear()
}

function statusIcon(status: string): string {
  if (status === 'success') return 'OK'
  if (status === 'failed') return 'FAIL'
  return 'WARN'
}

function toggleExpand(id: number) {
  expandedId.value = expandedId.value === id ? null : id
}

function fmtLatency(ms: number): string {
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}
</script>

<template>
  <div class="history-panel">
    <div class="history-toolbar">
      <span class="history-count">{{ t('history.total', { n: historyTotal }) }}</span>
      <div class="history-actions">
        <button class="btn btn-sm" @click="selectAll">
          {{ selectedIds.size === historyItems.length && historyItems.length > 0 ? t('history.deselectAll') : t('history.selectAll') }}
        </button>
        <button class="btn btn-sm" :disabled="selectedIds.size === 0" @click="handleDeleteSelected">
          {{ t('history.deleteSelected', { n: selectedIds.size }) }}
        </button>
        <button class="btn btn-sm" :disabled="historyItems.length === 0" @click="handleClearAll">
          {{ t('history.clearAll') }}
        </button>
        <button class="btn btn-sm" @click="loadHistory()">{{ t('history.refresh') }}</button>
      </div>
    </div>

    <div class="history-list" v-if="historyItems.length > 0">
      <div
        v-for="h in historyItems"
        :key="h.id"
        class="history-row"
        :class="{ selected: selectedIds.has(h.id) }"
      >
        <div class="history-row-main" @click="toggleExpand(h.id)">
          <input
            type="checkbox"
            :checked="selectedIds.has(h.id)"
            @click.stop="toggleSelect(h.id)"
          />
          <span class="history-status" :class="h.status">{{ statusIcon(h.status) }}</span>
          <span class="history-provider">{{ h.providerName }}</span>
          <span class="history-model">{{ h.model }}</span>
          <span class="protocol-badge">{{ h.protocol }}</span>
          <span class="history-latency">{{ fmtLatency(h.totalLatency) }}</span>
          <span class="history-time">{{ h.createdAt }}</span>
          <button class="btn-icon-sm" @click.stop="deleteHistoryItem(h.id)">&times;</button>
        </div>
        <div v-if="expandedId === h.id" class="history-detail">
          <div class="check-items">
            <div
              v-for="item in (h.results as CheckResult[])"
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

    <div v-else class="results-empty">
      <div class="icon">&#128203;</div>
      <div>{{ t('history.empty') }}</div>
    </div>
  </div>
</template>
