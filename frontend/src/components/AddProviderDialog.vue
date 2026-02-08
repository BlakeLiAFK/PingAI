<script setup lang="ts">
import { ref } from 'vue'
import type { ProtocolType } from '../types'
import { t } from '../i18n'
import { addProvider } from '../stores/check'

const emit = defineEmits<{ close: [] }>()

const form = ref({
  id: '',
  name: '',
  baseURL: '',
  protocol: 'openai' as ProtocolType,
  modelsText: '',
})

async function handleSubmit() {
  if (!form.value.id || !form.value.name || !form.value.baseURL) return
  const models = form.value.modelsText
    .split('\n')
    .map(s => s.trim())
    .filter(Boolean)
  await addProvider({
    id: form.value.id,
    name: form.value.name,
    baseURL: form.value.baseURL,
    protocol: form.value.protocol,
    models,
  })
  emit('close')
}
</script>

<template>
  <div class="dialog-overlay" @click.self="emit('close')">
    <div class="dialog">
      <div class="dialog-header">
        <h3>{{ t('addProvider.title') }}</h3>
        <button class="btn-icon" @click="emit('close')">&times;</button>
      </div>
      <div class="dialog-body">
        <div class="form-group">
          <label>{{ t('addProvider.id') }}</label>
          <input v-model="form.id" placeholder="my-provider" />
        </div>
        <div class="form-group">
          <label>{{ t('addProvider.name') }}</label>
          <input v-model="form.name" placeholder="My Provider" />
        </div>
        <div class="form-group">
          <label>{{ t('config.baseURL') }}</label>
          <input v-model="form.baseURL" placeholder="https://api.example.com/v1" />
        </div>
        <div class="form-group">
          <label>{{ t('config.protocol') }}</label>
          <select v-model="form.protocol">
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
            <option value="gemini">Gemini</option>
          </select>
        </div>
        <div class="form-group">
          <label>{{ t('addProvider.models') }}</label>
          <textarea v-model="form.modelsText" rows="4" placeholder="model-a&#10;model-b"></textarea>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="btn" @click="emit('close')">{{ t('addProvider.cancel') }}</button>
        <button class="btn btn-primary" @click="handleSubmit">{{ t('addProvider.add') }}</button>
      </div>
    </div>
  </div>
</template>
