<script setup lang="ts">
import { ref } from 'vue'
import { t } from '../i18n'
import { locale, setLocale } from '../i18n'
import type { Locale } from '../i18n'
import {
  providers,
  hiddenProviderIDs,
  setProviderVisibility,
  deleteProvider,
  resetAllProviders,
} from '../stores/check'
import AddProviderDialog from './AddProviderDialog.vue'

const emit = defineEmits<{ (e: 'close'): void }>()

const showAddDialog = ref(false)

function isVisible(id: string): boolean {
  return !hiddenProviderIDs.value.has(id)
}

async function toggleVisibility(id: string) {
  await setProviderVisibility(id, !isVisible(id))
}

async function handleDelete(id: string) {
  if (!confirm(t('settings.confirmDelete'))) return
  await deleteProvider(id)
}

async function handleResetAll() {
  if (!confirm(t('settings.confirmReset'))) return
  await resetAllProviders()
}

function handleLocaleChange(e: Event) {
  const val = (e.target as HTMLSelectElement).value as Locale
  setLocale(val)
}
</script>

<template>
  <div class="dialog-overlay" @click.self="emit('close')">
    <div class="dialog settings-dialog">
      <div class="dialog-header">
        <h3>{{ t('settings.title') }}</h3>
        <button class="btn-icon-sm" @click="emit('close')">&times;</button>
      </div>

      <div class="settings-body">
        <!-- 语言设置 -->
        <div class="settings-section">
          <div class="settings-section-header">
            <span class="settings-section-title">{{ t('settings.language') }}</span>
          </div>
          <div class="settings-lang-row">
            <select :value="locale" @change="handleLocaleChange">
              <option value="zh">中文</option>
              <option value="en">English</option>
            </select>
          </div>
        </div>

        <!-- 供应商管理 -->
        <div class="settings-section">
          <div class="settings-section-header">
            <span class="settings-section-title">{{ t('settings.providerManage') }}</span>
            <button class="btn btn-sm" @click="showAddDialog = true">+ {{ t('settings.add') }}</button>
          </div>

          <div class="settings-provider-list">
            <div
              v-for="p in providers"
              :key="p.id"
              class="settings-provider-row"
              :class="{ hidden: !isVisible(p.id) }"
            >
              <input
                type="checkbox"
                :checked="isVisible(p.id)"
                @change="toggleVisibility(p.id)"
              />
              <span class="settings-provider-name">{{ p.name }}</span>
              <span class="protocol-badge">{{ p.protocol }}</span>
              <span v-if="p.isBuiltin" class="settings-tag builtin">{{ t('settings.builtin') }}</span>
              <span v-else class="settings-tag custom">{{ t('settings.custom') }}</span>
              <button
                v-if="!p.isBuiltin"
                class="btn-icon-sm"
                @click="handleDelete(p.id)"
              >&times;</button>
            </div>
          </div>
        </div>
      </div>

      <div class="dialog-footer">
        <button class="btn btn-danger-outline" @click="handleResetAll">
          {{ t('settings.resetDefault') }}
        </button>
        <button class="btn" @click="emit('close')">{{ t('settings.close') }}</button>
      </div>
    </div>
  </div>

  <AddProviderDialog v-if="showAddDialog" @close="showAddDialog = false" />
</template>
