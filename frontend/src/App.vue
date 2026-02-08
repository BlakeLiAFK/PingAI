<script setup lang="ts">
import { onMounted } from 'vue'
import ProviderSidebar from './components/ProviderSidebar.vue'
import ConfigPanel from './components/ConfigPanel.vue'
import CheckCard from './components/CheckCard.vue'
import HistoryPanel from './components/HistoryPanel.vue'
import { t } from './i18n'
import {
  initProviders,
  isRunning,
  runAllChecks,
  exportReport,
  allResults,
  selectedProviderID,
  checkResults,
  activeView,
} from './stores/check'

onMounted(() => {
  initProviders()
})
</script>

<template>
  <div id="app">
    <header class="app-header">
      <div style="display:flex;align-items:baseline;">
        <h1>PingAI</h1>
        <span class="subtitle">AI API Availability Tester</span>
      </div>
      <div class="header-actions">
        <button class="btn" :disabled="allResults.length === 0" @click="exportReport">
          {{ t('app.export') }}
        </button>
        <button class="btn btn-primary" :disabled="isRunning" @click="runAllChecks">
          <span v-if="isRunning" class="spinner"></span>
          <span v-else>{{ t('app.batchCheck') }}</span>
        </button>
      </div>
    </header>

    <div class="main-content">
      <ProviderSidebar />
      <div class="content-area">
        <!-- 检测视图 -->
        <template v-if="activeView === 'check'">
          <ConfigPanel />
          <div class="results-area">
            <template v-if="checkResults.has(selectedProviderID)">
              <CheckCard :result="checkResults.get(selectedProviderID)!" />
            </template>
            <template v-for="r in allResults" :key="r.providerID">
              <CheckCard v-if="r.providerID !== selectedProviderID" :result="r" />
            </template>
            <div v-if="allResults.length === 0" class="results-empty">
              <div class="icon">&#9881;</div>
              <div>{{ t('app.emptyHint') }}</div>
              <div style="font-size:12px;">{{ t('app.emptySubHint') }}</div>
            </div>
          </div>
        </template>

        <!-- 历史记录视图 -->
        <template v-if="activeView === 'history'">
          <HistoryPanel />
        </template>
      </div>
    </div>
  </div>
</template>
