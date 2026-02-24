<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

function authHeaders () {
  const token = userStore.token
  if (!token) return {}
  return { Authorization: `Bearer ${token}` }
}

const instances = ref([])
const logs = ref([])
const logsLoading = ref(false)
const logsError = ref(null)
const logsNotConfigured = ref(false)
const logTypeFilter = ref('all')
const logLimit = ref(50)
const instanceIdFilter = ref('') // empty = all instances; otherwise instance id

const instanceNameById = computed(() => {
  const map = {}
  for (const i of instances.value) {
    map[i.id] = i.name || i.id
  }
  return map
})

const fetchInstances = async () => {
  try {
    const response = await axios.get('/api/v1/instances', { headers: authHeaders() })
    instances.value = response.data || []
  } catch (err) {
    console.error('Failed to load instances', err)
    instances.value = []
  }
}

const fetchLogs = async () => {
  logsError.value = null
  logsLoading.value = true
  try {
    const params = new URLSearchParams()
    if (logTypeFilter.value && logTypeFilter.value !== 'all') params.set('type', logTypeFilter.value)
    if (logLimit.value > 0) params.set('limit', String(logLimit.value))
    if (instanceIdFilter.value) params.set('instanceId', instanceIdFilter.value)
    const url = `/api/v1/logs${params.toString() ? '?' + params.toString() : ''}`
    const response = await axios.get(url, { headers: authHeaders() })
    logs.value = response.data ?? []
    logsNotConfigured.value = false
  } catch (err) {
    const status = err.response?.status
    const msg = err.response?.data?.error ?? err.message ?? 'Failed to load logs'
    logs.value = []
    if (status === 503) {
      logsNotConfigured.value = true
      logsError.value = 'Logs not configured for this environment.'
    } else {
      logsNotConfigured.value = false
      logsError.value = msg
    }
  } finally {
    logsLoading.value = false
  }
}

function formatLogTime (ts) {
  if (!ts) return '—'
  try {
    const d = new Date(ts)
    return isNaN(d.getTime()) ? ts : d.toLocaleString()
  } catch {
    return ts
  }
}

function formatLogDetails (details) {
  if (details == null || (Array.isArray(details) && details.length === 0)) return ''
  try {
    const o = typeof details === 'string' ? JSON.parse(details) : details
    return JSON.stringify(o)
  } catch {
    return String(details)
  }
}

function instanceDisplayName (entry) {
  return instanceNameById.value[entry.instanceId] ?? entry.instanceId ?? '—'
}

onMounted(() => {
  fetchInstances().then(() => fetchLogs())
})
</script>

<template>
  <div class="container-fluid">
    <h1 class="h3 mb-4">Activity &amp; Logs</h1>
    <p class="text-secondary small mb-4">
      Audit logs (your actions) and service logs (e.g. status changes) across all instances. Filter by instance name, type, and limit.
    </p>

    <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
      <label class="text-muted small mb-0">Instance</label>
      <select v-model="instanceIdFilter" class="form-select form-select-sm" style="width: auto; max-width: 12rem" @change="fetchLogs">
        <option value="">All instances</option>
        <option v-for="inst in instances" :key="inst.id" :value="inst.id">{{ inst.name || inst.id }}</option>
      </select>
      <label class="text-muted small mb-0">Type</label>
      <select v-model="logTypeFilter" class="form-select form-select-sm" style="width: auto" @change="fetchLogs">
        <option value="all">All logs</option>
        <option value="audit">Audit only</option>
        <option value="service">Service only</option>
      </select>
      <input v-model.number="logLimit" type="number" class="form-control form-control-sm" style="width: 5rem" min="1" max="500" placeholder="Limit" @keydown.enter="fetchLogs">
      <span class="text-muted small">entries</span>
      <button type="button" class="btn btn-outline-secondary btn-sm" :disabled="logsLoading" @click="fetchLogs">
        <i class="bi bi-arrow-clockwise"></i> {{ logsLoading ? 'Loading…' : 'Refresh' }}
      </button>
    </div>

    <p v-if="logsError" class="small mb-2" :class="logsNotConfigured ? 'text-muted' : 'text-danger'">{{ logsError }}</p>
    <div v-else-if="logsLoading && logs.length === 0" class="text-center py-4 text-muted small">
      Loading logs…
    </div>
    <div v-else-if="logs.length === 0" class="text-muted small py-3">
      No log entries yet.
    </div>
    <div v-else class="table-responsive">
      <table class="table table-sm table-hover mb-0">
        <thead class="table-light">
          <tr>
            <th class="text-nowrap">Time</th>
            <th class="text-nowrap">Type</th>
            <th class="text-nowrap">Instance</th>
            <th class="text-nowrap">Action</th>
            <th>Message / details</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="entry in logs" :key="entry.id">
            <td class="text-nowrap small">{{ formatLogTime(entry.timestamp) }}</td>
            <td>
              <span class="badge" :class="entry.type === 'audit' ? 'bg-primary' : 'bg-secondary'">
                {{ entry.type }}
              </span>
            </td>
            <td class="small">{{ instanceDisplayName(entry) }}</td>
            <td class="font-monospace small">{{ entry.action }}</td>
            <td class="small">
              <span v-if="entry.message">{{ entry.message }}</span>
              <code v-else-if="formatLogDetails(entry.details)" class="d-block mt-1 p-1 bg-light rounded small text-break">{{ formatLogDetails(entry.details) }}</code>
              <span v-else-if="formatLogDetails(entry.metadata)" class="text-muted">{{ formatLogDetails(entry.metadata) }}</span>
              <span v-else class="text-muted">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
