<script setup>
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { Modal } from 'bootstrap'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()
const instances = ref([])
const loading = ref(true)
const error = ref(null)

const createModalRef = ref(null)
let modalInstance = null
const isCreating = ref(null)

const newInstance = reactive({
  name: '',
  capacity: '1Gi',
  redisReplicas: 3,
  sentinelReplicas: 3
})

const fetchInstances = async () => {
  try {
    const response = await axios.get('/api/v1/instances', {
      headers: {
        'X-User': userStore.username
      }
    })
    instances.value = response.data || []
    loading.value = false
  } catch (err) {
    console.error('API Error:', err)
    error.value = 'Failed to load instances. Is the backend running?'
    loading.value = false
  }
}

const openCreateModal = () => {
  newInstance.name = ''
  newInstance.capacity = '1Gi'
  newInstance.redisReplicas = 3
  newInstance.sentinelReplicas = 3

  modalInstance.show()
}

const createInstance = async () => {
  isCreating.value = true
  try {
    const payload = {
      ...newInstance,
      name: newInstance.name.toLowerCase()
    }

    await axios.post('/api/v1/instances', payload, {
      headers: {
        'X-User': userStore.username
      }
    })

    modalInstance.hide()
    await fetchInstances()
  } catch (err) {
    alert('Failed to create instance: ' + (err.response?.data?.error || err.message))
  } finally {
    isCreating.value = false
  }
}

const deleteInstance = async (instance) => {
  if (!confirm(`Are you sure you want to delete ${instance.name}? This cannot be undone.`)) {
    return
  }

  try {
    await axios.delete(`/api/v1/instances/${instance.id}`, {
      headers: {
        'X-User': userStore.username
      }
    })
    await fetchInstances()
  } catch (err) {
    alert('Failed to delete instance: ' + (err.response?.data?.error || err.message))
  }
}


onMounted(() => {
  fetchInstances()
  modalInstance = new Modal(createModalRef.value)
})
</script>

<template>
  <div class="container mt-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h1>Redis Instances</h1>

      <div>
        <button @click="openCreateModal" class="btn btn-primary me-2">
          + Create Instance
        </button>
        <button @click="fetchInstances" class="btn btn-outline-secondary btn-sm" title="Refresh List">
          <i class="bi bi-arrow-clockwise"></i>
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading data...</span>
      </div>
      <p class="mt-2 text-muted">Fetching your Redis instances...</p>
    </div>

    <div v-else-if="error" class="alert alert-danger shadow-sm">
      <h4 class="alert-heading">Error</h4>
      <p>{{ error }}</p>
      <hr>
      <button @click="fetchInstances" class="btn btn-outline-danger btn-sm">Try Again</button>
    </div>

    <div v-else class="row">
      <div v-for="instance in instances" :key="instance.id" class="col-md-4 mb-4">
        <div class="card shadow-sm h-100">
          <div class="card-header d-flex justify-content-between align-items-center">
            <strong>{{ instance.name }}</strong>
            <span class="badge"
              :class="instance.status === 'running' ? 'bg-success' : 'bg-warning text-dark'">
              {{ instance.status }}
            </span>
          </div>
          <div class="card-body">
            <p class="card-text">
              <strong>Replicas:</strong> {{ instance.redisReplicas }}<br>
              <strong>Capacity:</strong> {{ instance.capacity }}<br>
              <strong>Namespace:</strong> {{ instance.namespace }}
            </p>
            <button class="btn btn-outline-primary btn-sm" @click="router.push({ name: 'instance-detail', params: { id: instance.id } })">Manage</button>
            <button class="btn btn-outline-danger btn-sm float-end ms-2" @click="deleteInstance(instance)">
              <i class="bi bi-trash"></i> Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" ref="createModalRef" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Create New Redis Instance</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="createInstance">
              <div class="mb-3">
                <label class="form-label">Instance Name</label>
                <input v-model="newInstance.name" type="text" class="form-control" placeholder="e.g. redis-cache" pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?" title="Lowercase alphanumeric characters only (e.g. 'my-redis')" required>
                <div class="form-text">Lowercase alphanumeric characters only (e.g. 'my-redis')</div>
              </div>
              <div class="mb-3">
                <label class="form-label">Capacity</label>
                <input v-model="newInstance.capacity" type="text" class="form-control" placeholder="e.g. 10Gi" required>
              </div>
              <div class="row">
                <div class="col-6 mb-3">
                  <label class="form-label">Redis Replicas</label>
                  <input v-model.number="newInstance.redisReplicas" type="number" class="form-control" min="1" required>
                </div>
                <div class="col-6 mb-3">
                  <label class="form-label">Sentinel Replicas</label>
                  <input v-model.number="newInstance.sentinelReplicas" type="number" class="form-control" min="1" required>
                </div>
              </div>

              <div class="d-flex justify-content-end mt-4">
                <button type="button" class="btn btn-secondary me-2" data-bs-dismiss="modal">Cancel</button>
                <button type="submit" class="btn btn-primary" :disabled="isCreating">
                  <span v-if="isCreating" class="spinner-border spinner-border-sm me-1"></span>
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!loading && !error && instances.length === 0" class="alert alert-warning">
      No Redis instances found.
    </div>
  </div>
</template>

