<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const username = ref('')
const password = ref('')
const error = ref('')
const router = useRouter()
const userStore = useUserStore()

const handleLogin = async () => {
  error.value = ''
  try {
    const success = await userStore.login(username.value, password.value)
    if (success) {
      router.push('/')
    }
  } catch (err) {
    if (err.response && err.response.status === 401) {
      error.value = 'Invalid username or password.'
    } else {
      error.value = 'An error occurred during login.'
    }
  }
}
</script>

<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6">
        <div class="card shadow">
          <div class="card-header bg-primary text-white">Login</div>
          <div class="card-body">
            <form @submit.prevent="handleLogin">
              <div class="mb-3">
                <label for="username" class="form-label">Username</label>
                <input
                  v-model="username"
                  type="text"
                  class="form-control"
                  id="username"
                  required
                >
              </div>
              <div class="mb-3">
                <label for="password" class="form-label">Password</label>
                <input
                  v-model="password"
                  type="password"
                  class="form-control"
                  id="password"
                  required
                >
              </div>
              <div v-if="error" class="alert alert-danger">
                {{ error }}
              </div>
              <button type="submit" class="btn btn-primary w-100">Login</button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
