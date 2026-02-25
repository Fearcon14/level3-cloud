<script setup>
import { ref, onMounted, computed, watch } from 'vue'
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
const instanceIdFilter = ref('') // empty = all instances; otherwise instance id
const pageSize = ref(20) // 10 or 20 entries per page
const currentPage = ref(1)
const totalCount = ref(0) // from X-Total-Count header

const instanceNameById = computed(() => {
  const map = {}
  for (const i of instances.value) {
    map[i.id] = i.name || i.id
  }
  return map
})

const totalPages = computed(() => {
  const total = totalCount.value
  const size = pageSize.value
  if (size <= 0 || total <= 0) return 1
  return Math.max(1, Math.ceil(total / size))
})

const pageNumbers = computed(() => {
  const total = totalPages.value
  const current = currentPage.value
  const delta = 2
  const range = []
  const rangeWithDots = []
  let l
  for (let i = 1; i <= total; i++) {
    if (i === 1 || i === total || (i >= current - delta && i <= current + delta)) {
      range.push(i)
    }
  }
  for (const i of range) {
    if (l !== undefined && i - l !== 1) {
      rangeWithDots.push(-1) // ellipsis
    }
    rangeWithDots.push(i)
    l = i
  }
  return rangeWithDots
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
    params.set('limit', String(pageSize.value))
    params.set('offset', String((currentPage.value - 1) * pageSize.value))
    if (instanceIdFilter.value) params.set('instanceId', instanceIdFilter.value)
    const url = `/api/v1/logs?${params.toString()}`
    const response = await axios.get(url, { headers: authHeaders() })
    logs.value = response.data ?? []
    const countHeader = response.headers?.['x-total-count']
    totalCount.value = countHeader ? parseInt(countHeader, 10) || 0 : logs.value.length
    logsNotConfigured.value = false
  } catch (err) {
    const status = err.response?.status
    const msg = err.response?.data?.error ?? err.message ?? 'Failed to load logs'
    logs.value = []
    totalCount.value = 0
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

function goToPage (page) {
  const p = Math.max(1, Math.min(page, totalPages.value))
  if (p !== currentPage.value) {
    currentPage.value = p
    fetchLogs()
  }
}

// When filters or page size change, reset to page 1 and refetch
watch([instanceIdFilter, logTypeFilter, pageSize], () => {
  currentPage.value = 1
  fetchLogs()
})

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
      Audit logs (your actions) and service logs (e.g. status changes) across all instances. Filter by instance, type, and use the pager to browse.
    </p>

    <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
      <label class="text-muted small mb-0">Instance</label>
      <select v-model="instanceIdFilter" class="form-select form-select-sm" style="width: auto; max-width: 12rem">
        <option value="">All instances</option>
        <option v-for="inst in instances" :key="inst.id" :value="inst.id">{{ inst.name || inst.id }}</option>
      </select>
      <label class="text-muted small mb-0">Type</label>
      <select v-model="logTypeFilter" class="form-select form-select-sm" style="width: auto">
        <option value="all">All logs</option>
        <option value="audit">Audit only</option>
        <option value="service">Service only</option>
      </select>
      <label class="text-muted small mb-0">Per page</label>
      <select v-model.number="pageSize" class="form-select form-select-sm" style="width: auto">
        <option :value="10">10</option>
        <option :value="20">20</option>
      </select>
      <button type="button" class="btn btn-outline-secondary btn-sm ms-2" :disabled="logsLoading" @click="fetchLogs">
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

    <nav v-if="totalPages > 1 && !logsLoading && logs.length > 0" class="d-flex justify-content-between align-items-center flex-wrap gap-2 mt-3" aria-label="Logs pagination">
      <div class="text-muted small">
        Page {{ currentPage }} of {{ totalPages }} ({{ totalCount }} entries)
      </div>
      <ul class="pagination pagination-sm mb-0">
        <li class="page-item" :class="{ disabled: currentPage <= 1 }">
          <button type="button" class="page-link" :disabled="currentPage <= 1" aria-label="Previous" @click="goToPage(currentPage - 1)">
            <i class="bi bi-chevron-left"></i>
          </button>
        </li>
        <li v-for="p in pageNumbers" :key="p" class="page-item" :class="{ active: p === currentPage, disabled: p === -1 }">
          <button v-if="p === -1" type="button" class="page-link" disabled>…</button>
          <button v-else type="button" class="page-link" @click="goToPage(p)">{{ p }}</button>
        </li>
        <li class="page-item" :class="{ disabled: currentPage >= totalPages }">
          <button type="button" class="page-link" :disabled="currentPage >= totalPages" aria-label="Next" @click="goToPage(currentPage + 1)">
            <i class="bi bi-chevron-right"></i>
          </button>
        </li>
      </ul>
    </nav>
  </div>
</template>
