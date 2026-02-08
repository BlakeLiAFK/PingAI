<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ProtocolType } from '../types'
import { t } from '../i18n'
import {
  providers,
  selectedProviderID,
  checkConfigs,
  isRunning,
  isBatchRunning,
  runSingleCheck,
  autoSaveConfig,
  resetProviderConfig,
} from '../stores/check'
import BatchKeyDialog from './BatchKeyDialog.vue'

const showBatchDialog = ref(false)

const currentProvider = computed(() =>
  providers.value.find(p => p.id === selectedProviderID.value)
)

const config = computed(() => checkConfigs.get(selectedProviderID.value))

const isBuiltin = computed(() => currentProvider.value?.isBuiltin ?? false)

function updateConfig(field: string, value: string) {
  const cfg = checkConfigs.get(selectedProviderID.value)
  if (cfg) {
    (cfg as any)[field] = value
    autoSaveConfig(selectedProviderID.value)
  }
}

async function runCheck() {
  await runSingleCheck(selectedProviderID.value)
}

async function handleReset() {
  await resetProviderConfig(selectedProviderID.value)
}
</script>

<template>
  <div class="config-panel" v-if="config">
    <div class="config-row">
      <div class="form-group fg-url">
        <label>{{ t('config.baseURL') }}</label>
        <input
          type="text"
          :value="config.baseURL"
          @input="updateConfig('baseURL', ($event.target as HTMLInputElement).value)"
          placeholder="https://api.example.com/v1"
        />
      </div>
      <div class="form-group fg-key">
        <label>{{ t('config.apiKey') }}</label>
        <input
          type="password"
          :value="config.apiKey"
          @input="updateConfig('apiKey', ($event.target as HTMLInputElement).value)"
          placeholder="sk-..."
        />
      </div>
      <div class="form-group" style="min-width:110px;">
        <label>{{ t('config.protocol') }}</label>
        <select
          :value="config.protocol"
          @change="updateConfig('protocol', ($event.target as HTMLSelectElement).value)"
        >
          <option value="openai">OpenAI</option>
          <option value="anthropic">Anthropic</option>
          <option value="gemini">Gemini</option>
        </select>
      </div>
      <div class="form-group fg-model">
        <label>{{ t('config.model') }}</label>
        <template v-if="currentProvider && currentProvider.models?.length > 0">
          <select
            :value="config.model"
            @change="updateConfig('model', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="m in currentProvider.models" :key="m" :value="m">{{ m }}</option>
          </select>
        </template>
        <template v-else>
          <input
            type="text"
            :value="config.model"
            @input="updateConfig('model', ($event.target as HTMLInputElement).value)"
            placeholder="model-name"
          />
        </template>
      </div>
      <button
        v-if="isBuiltin"
        class="btn btn-reset"
        @click="handleReset"
        :title="t('config.reset')"
      >
        {{ t('config.reset') }}
      </button>
      <button
        class="btn"
        :disabled="isRunning || isBatchRunning || !config.model"
        @click="showBatchDialog = true"
        :title="t('config.batch')"
      >
        {{ t('config.batch') }}
      </button>
      <button
        class="btn btn-primary"
        :disabled="isRunning || !config.model"
        @click="runCheck"
      >
        <span v-if="isRunning" class="spinner"></span>
        <span v-else>{{ t('config.check') }}</span>
      </button>
    </div>
  </div>

  <BatchKeyDialog v-if="showBatchDialog" @close="showBatchDialog = false" />
</template>
