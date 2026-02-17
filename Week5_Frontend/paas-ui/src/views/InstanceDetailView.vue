<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'

const route = useRoute()
const router = useRouter()

const instance = ref(null)
const loading = ref(true)
const error = ref(null)
const showPassword = ref(false)

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    // Optional: could add a toast notification here
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}

const fetchInstance = async () => {
  const instanceId = route.params.id
  try {
    const response = await axios.get(`/api/v1/instances/${instanceId}`)
    instance.value = response.data
    loading.value = false
  } catch (err) {
    console.error('API Error:', err)
    error.value = 'Failed to load instance details.'
    loading.value = false
  }
}

onMounted(() => {
  fetchInstance()
})
</script>

<template>
  <div class="container mt-5">
    <div class="mb-4">
      <button @click="router.push('/')" class="btn btn-outline-secondary btn-sm">
        &larr; Back to Dashboard
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
      <div class="card-header bg-light d-flex justify-content-between align-items-center">
        <h2 class="h4 mb-0">{{ instance.name }}</h2>
        <span class="badge"
          :class="instance.status === 'running' ? 'bg-success' : 'bg-warning text-dark'">
          {{ instance.status }}
        </span>
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
      </div>
    </div>
  </div>
</template>
