<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const instances = ref([])
const loading = ref(true)
const error = ref(null)

const fetchInstances = async () => {
  try {
    const response = await axios.get('/api/v1/instances', {
      headers: {
        'X-User': userStore.username
      }
    })
    instances.value = response.data
    loading.value = false
  } catch (err) {
    console.error('API Error:', err)
    error.value = 'Failed to load instances. Is the backend running?'
    loading.value = false
  }
}

onMounted(() => {
  fetchInstances()
})
</script>

<template>
  <div class="container mt-5">
    <h1>Redis Instances</h1>

    <div v-if="loading" class="alert alert-info">
      Loading data...
    </div>

    <div v-else-if="error" class="alert alert-danger">
      {{ error }}
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
            <button class="btn btn-outline-primary btn-sm">Manage</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!loading && !error && instances.length === 0" class="alert alert-warning">
      No Redis instances found.
    </div>
  </div>
</template>


<!-- <p v-for="instance in instances" :key="instance.id">
  {{ instance.name }} - {{ instance.status }} - {{ instance.capacity }}
</p>
<script>
export default {
  data() {
    return {
      instances: [],
    }
  }
}
</script> -->
