<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { Modal } from 'bootstrap'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

function authHeaders () {
  const token = userStore.token
  if (!token) return {}
  return { Authorization: `Bearer ${token}` }
}

const instance = ref(null)
const loading = ref(true)
const error = ref(null)
const showPassword = ref(false)

const editModalRef = ref(null)
let editModalInstance = null
const isSaving = ref(false)
const editForm = reactive({
  name: '',
  capacity: '',
  redisReplicas: 0,
  sentinelReplicas: 0
})

// Cache API: endpoints and state
const cacheGetKey = ref('')
const cacheGetResult = ref(null)
const cacheGetError = ref(null)
const cacheGetLoading = ref(false)
const cacheSetKey = ref('')
const cacheSetValue = ref('')
const cacheSetTtl = ref(0)
const cacheSetLoading = ref(false)
const cacheSetError = ref(null)
const cacheSetSuccess = ref(false)

function cacheGetEndpoint () {
  const id = instance.value?.id ?? ':id'
  return `/api/v1/instances/${id}/cache/:key`
}
function cachePostEndpoint () {
  const id = instance.value?.id ?? ':id'
  return `/api/v1/instances/${id}/cache`
}

const fetchCacheValue = async () => {
  const key = (cacheGetKey.value || '').trim()
  if (!key) {
    cacheGetError.value = 'Enter a key'
    return
  }
  cacheGetError.value = null
  cacheGetResult.value = null
  cacheGetLoading.value = true
  try {
    const response = await axios.get(`/api/v1/instances/${instance.value.id}/cache/${encodeURIComponent(key)}`, { headers: authHeaders() })
    cacheGetResult.value = response.data
  } catch (err) {
    cacheGetError.value = err.response?.data?.error ?? err.message ?? 'Failed to get value'
    cacheGetResult.value = null
  } finally {
    cacheGetLoading.value = false
  }
}

const setCacheValue = async () => {
  const key = (cacheSetKey.value || '').trim()
  if (!key) {
    cacheSetError.value = 'Key is required'
    return
  }
  cacheSetError.value = null
  cacheSetSuccess.value = false
  cacheSetLoading.value = true
  try {
    const payload = { key, value: cacheSetValue.value }
    if (cacheSetTtl.value > 0) payload.ttlSeconds = cacheSetTtl.value
    await axios.post(`/api/v1/instances/${instance.value.id}/cache`, payload, { headers: authHeaders() })
    cacheSetSuccess.value = true
  } catch (err) {
    cacheSetError.value = err.response?.data?.error ?? err.message ?? 'Failed to set value'
  } finally {
    cacheSetLoading.value = false
  }
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}

const fetchInstance = async () => {
  const instanceId = route.params.id
  try {
    const response = await axios.get(`/api/v1/instances/${instanceId}`, { headers: authHeaders() })
    instance.value = response.data
    error.value = null
  } catch (err) {
    console.error('API Error:', err)
    error.value = 'Failed to load instance details.'
  } finally {
    loading.value = false
  }
}

function openEditModal () {
  if (!instance.value) return
  editForm.name = instance.value.name ?? ''
  editForm.capacity = instance.value.capacity ?? ''
  editForm.redisReplicas = instance.value.redisReplicas ?? 0
  editForm.sentinelReplicas = instance.value.sentinelReplicas ?? 0
  editModalInstance.show()
}

function buildPatchPayload () {
  if (!instance.value) return {}
  const payload = {}
  if (editForm.name !== (instance.value.name ?? '')) payload.name = editForm.name
  if (String(editForm.capacity).trim() !== String(instance.value.capacity ?? '').trim()) payload.capacity = String(editForm.capacity).trim()
  if (Number(editForm.redisReplicas) !== Number(instance.value.redisReplicas ?? 0)) payload.redisReplicas = Number(editForm.redisReplicas)
  if (Number(editForm.sentinelReplicas) !== Number(instance.value.sentinelReplicas ?? 0)) payload.sentinelReplicas = Number(editForm.sentinelReplicas)
  return payload
}

const submitEdit = async () => {
  const payload = buildPatchPayload()
  if (Object.keys(payload).length === 0) {
    alert('No changes detected. Edit at least one field to update the instance.')
    return
  }
  isSaving.value = true
  try {
    await axios.patch(`/api/v1/instances/${instance.value.id}`, payload, { headers: authHeaders() })
    editModalInstance.hide()
    await fetchInstance()
  } catch (err) {
    alert('Failed to update instance: ' + (err.response?.data?.error || err.message))
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  fetchInstance()
  editModalInstance = new Modal(editModalRef.value)
})
</script>

<template>
  <div class="container mt-5">
    <div class="mb-4 d-flex justify-content-between align-items-center">
      <button @click="router.push('/')" class="btn btn-outline-secondary btn-sm">
        &larr; Back to Dashboard
      </button>
      <button
        v-if="!loading && (instance || error)"
        @click="loading = true; fetchInstance()"
        class="btn btn-outline-secondary btn-sm"
        title="Refresh"
      >
        <i class="bi bi-arrow-clockwise"></i>
      </button>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading details...</span>
      </div>
      <p class="mt-2 text-muted">Fetching instance details...</p>
    </div>

    <div v-else-if="error" class="alert alert-danger shadow-sm">
      <h4 class="alert-heading">Error</h4>
      <p>{{ error }}</p>
      <hr>
      <button @click="fetchInstance" class="btn btn-outline-danger btn-sm">Try Again</button>
    </div>

    <div v-else-if="instance" class="card shadow-sm">
      <div class="card-header bg-light d-flex justify-content-between align-items-center flex-wrap gap-2">
        <h2 class="h4 mb-0">{{ instance.name }}</h2>
        <div class="d-flex align-items-center gap-2">
          <button type="button" class="btn btn-outline-primary btn-sm" title="Modify" @click="openEditModal">
            <i class="bi bi-pencil"></i>
          </button>
          <span class="badge"
            :class="instance.status === 'running' ? 'bg-success' : 'bg-warning text-dark'">
            {{ instance.status }}
          </span>
        </div>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-6 mb-3">
            <h5 class="text-muted text-uppercase fs-6 fw-bold">Configuration</h5>
            <ul class="list-group list-group-flush">
              <li class="list-group-item d-flex justify-content-between">
                <span>ID</span>
                <span class="font-monospace">{{ instance.id }}</span>
              </li>
              <li class="list-group-item d-flex justify-content-between">
                <span>Capacity</span>
                <span>{{ instance.capacity }}</span>
              </li>
              <li class="list-group-item d-flex justify-content-between">
                <span>Redis Replicas</span>
                <span>{{ instance.redisReplicas }}</span>
              </li>
              <li class="list-group-item d-flex justify-content-between">
                <span>Sentinel Replicas</span>
                <span>{{ instance.sentinelReplicas }}</span>
              </li>
            </ul>
          </div>

          <div class="col-md-6 mb-3">
            <h5 class="text-muted text-uppercase fs-6 fw-bold">Kubernetes Info</h5>
            <ul class="list-group list-group-flush">
              <li class="list-group-item d-flex justify-content-between">
                <span>Namespace</span>
                <span>{{ instance.namespace }}</span>
              </li>
              <li class="list-group-item d-flex justify-content-between">
                <span>Service Name</span>
                <span>{{ instance.publicServiceName }}</span>
              </li>
              <li class="list-group-item d-flex justify-content-between">
                <span>Cluster Hostname</span>
                <span class="text-truncate ms-2" :title="instance.publicHostname" style="max-width: 200px;">
                  {{ instance.publicHostname }}
                </span>
              </li>
            </ul>
          </div>
        </div>

        <div class="mt-4 pt-3 border-top">
          <h5 class="text-muted text-uppercase fs-6 fw-bold mb-3">Connection Info</h5>
          
          <div class="row">
            <div class="col-md-8">
              <label class="form-label text-secondary small">Public Endpoint</label>
              <div class="input-group mb-3">
                <input type="text" class="form-control font-monospace" :value="instance.publicEndpoint" readonly>
                <button class="btn btn-outline-secondary" type="button" @click="copyToClipboard(instance.publicEndpoint)">
                  <i class="bi bi-clipboard"></i> Copy
                </button>
              </div>
            </div>
            
            <div class="col-md-4">
              <label class="form-label text-secondary small">Port</label>
              <input type="text" class="form-control font-monospace" :value="instance.publicPort" readonly>
            </div>
          </div>

          <div class="mb-3">
            <label class="form-label text-secondary small d-flex justify-content-between">
              <span>Password</span>
              <a href="#" @click.prevent="showPassword = !showPassword" class="text-decoration-none small">
                {{ showPassword ? 'Hide' : 'Show' }}
              </a>
            </label>
            <div class="input-group">
              <input 
                :type="showPassword ? 'text' : 'password'" 
                class="form-control font-monospace" 
                :value="instance.password" 
                readonly
              >
              <button class="btn btn-outline-secondary" type="button" @click="copyToClipboard(instance.password)">
                <i class="bi bi-clipboard"></i> Copy
              </button>
            </div>
          </div>
        </div>

        <!-- Cache API endpoints and try-it UI -->
        <div class="mt-4 pt-3 border-top">
          <h5 class="text-muted text-uppercase fs-6 fw-bold mb-3">Cache API</h5>
          <p class="text-secondary small mb-3">
            Use these API endpoints to read and write cache entries for this instance.
          </p>
          <ul class="list-group list-group-flush mb-4">
            <li class="list-group-item d-flex align-items-start gap-2">
              <span class="badge bg-primary">GET</span>
              <code class="font-monospace small flex-grow-1" :title="cacheGetEndpoint()">{{ cacheGetEndpoint() }}</code>
              <button type="button" class="btn btn-outline-secondary btn-sm" @click="copyToClipboard(cacheGetEndpoint())">
                <i class="bi bi-clipboard"></i>
              </button>
            </li>
            <li class="list-group-item d-flex align-items-start gap-2">
              <span class="badge bg-success">POST</span>
              <code class="font-monospace small flex-grow-1" :title="cachePostEndpoint()">{{ cachePostEndpoint() }}</code>
              <button type="button" class="btn btn-outline-secondary btn-sm" @click="copyToClipboard(cachePostEndpoint())">
                <i class="bi bi-clipboard"></i>
              </button>
            </li>
          </ul>

          <div class="row">
            <div class="col-md-6 mb-3">
              <h6 class="text-muted small fw-bold mb-2">Get value</h6>
              <div class="input-group mb-2">
                <input v-model="cacheGetKey" type="text" class="form-control font-monospace" placeholder="Cache key" @keydown.enter.prevent="fetchCacheValue">
                <button type="button" class="btn btn-outline-primary" :disabled="cacheGetLoading" @click="fetchCacheValue">
                  {{ cacheGetLoading ? '…' : 'GET' }}
                </button>
              </div>
              <p v-if="cacheGetError" class="text-danger small mb-0">{{ cacheGetError }}</p>
              <div v-if="cacheGetResult" class="small">
                <span class="text-muted">Value:</span>
                <code class="d-block mt-1 p-2 bg-light rounded">{{ cacheGetResult.value }}</code>
              </div>
            </div>
            <div class="col-md-6 mb-3">
              <h6 class="text-muted small fw-bold mb-2">Set value</h6>
              <div class="mb-2">
                <input v-model="cacheSetKey" type="text" class="form-control form-control-sm font-monospace mb-1" placeholder="Key">
                <input v-model="cacheSetValue" type="text" class="form-control form-control-sm font-monospace mb-1" placeholder="Value">
                <input v-model.number="cacheSetTtl" type="number" class="form-control form-control-sm font-monospace" placeholder="TTL (seconds, 0 = no expiry)" min="0">
              </div>
              <button type="button" class="btn btn-outline-success btn-sm" :disabled="cacheSetLoading" @click="setCacheValue">
                {{ cacheSetLoading ? '…' : 'POST' }}
              </button>
              <p v-if="cacheSetError" class="text-danger small mb-0 mt-1">{{ cacheSetError }}</p>
              <p v-if="cacheSetSuccess" class="text-success small mb-0 mt-1">Value stored.</p>
            </div>
          </div>
        </div>

      </div>
    </div>

    <!-- Modify instance modal -->
    <div ref="editModalRef" class="modal fade" tabindex="-1" aria-labelledby="editModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="editModalLabel">Modify instance</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <p class="text-muted small">Change only the fields you want to update. Unchanged values are left as-is.</p>
            <form id="edit-instance-form" @submit.prevent="submitEdit">
              <div class="mb-3">
                <label class="form-label">Display name</label>
                <input v-model="editForm.name" type="text" class="form-control" placeholder="e.g. my-redis">
              </div>
              <div class="mb-3">
                <label class="form-label">Persistent volume size (capacity)</label>
                <input v-model="editForm.capacity" type="text" class="form-control" placeholder="e.g. 10Gi">
              </div>
              <div class="row">
                <div class="col-6 mb-3">
                  <label class="form-label">Redis replicas</label>
                  <input v-model.number="editForm.redisReplicas" type="number" class="form-control" min="1" max="9">
                </div>
                <div class="col-6 mb-3">
                  <label class="form-label">Sentinel replicas</label>
                  <input v-model.number="editForm.sentinelReplicas" type="number" class="form-control" min="1" max="9">
                </div>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
            <button type="submit" form="edit-instance-form" class="btn btn-primary" :disabled="isSaving">
              {{ isSaving ? 'Saving…' : 'Save changes' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
