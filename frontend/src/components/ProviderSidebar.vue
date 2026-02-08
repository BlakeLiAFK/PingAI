<script setup lang="ts">
import { ref } from 'vue'
import { t } from '../i18n'
import {
  visibleProviders,
  selectedProviderID,
  checkConfigs,
  checkResults,
  activeView,
} from '../stores/check'
import SettingsDialog from './SettingsDialog.vue'

const showSettings = ref(false)

function selectProvider(id: string) {
  selectedProviderID.value = id
  activeView.value = 'check'
}

function getDotClass(id: string): string {
  const result = checkResults.get(id)
  if (result) {
    const hasFail = result.results?.some(r => r.status === 'failed')
    return hasFail ? 'checked-fail' : 'checked-ok'
  }
  const cfg = checkConfigs.get(id)
  if (cfg && cfg.apiKey) return 'configured'
  return ''
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <span class="sidebar-title">{{ t('sidebar.providers') }}</span>
      <button class="btn-icon" @click="showSettings = true" :title="t('settings.title')">&#9881;</button>
    </div>
    <div class="provider-list">
      <div
        v-for="p in visibleProviders"
        :key="p.id"
        class="provider-item"
        :class="{ active: selectedProviderID === p.id && activeView === 'check' }"
        @click="selectProvider(p.id)"
      >
        <span class="dot" :class="getDotClass(p.id)"></span>
        <span class="provider-name">{{ p.name }}</span>
        <span class="protocol-badge">{{ p.protocol }}</span>
      </div>
    </div>
    <div class="sidebar-footer">
      <div
        class="provider-item"
        :class="{ active: activeView === 'history' }"
        @click="activeView = 'history'"
      >
        <span class="dot"></span>
        <span>{{ t('sidebar.history') }}</span>
      </div>
    </div>
    <SettingsDialog v-if="showSettings" @close="showSettings = false" />
  </aside>
</template>
