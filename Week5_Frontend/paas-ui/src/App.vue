<script setup>
import { RouterView, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const router = useRouter()

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}
</script>

<template>
  <header v-if="userStore.username">
    <nav class="navbar navbar-expand-lg navbar-light bg-light mb-4">
      <div class="container-fluid">
        <router-link to="/" class="navbar-brand text-decoration-none">PaaS Dashboard</router-link>
        <div class="d-flex align-items-center">
          <span class="me-3">Hello, <strong>{{ userStore.username }}</strong></span>
          <router-link to="/logs" class="btn btn-outline-secondary btn-sm me-2">Activity &amp; Logs</router-link>
          <a
            href="https://grafana.kevin-sinn.runs.onstackit.cloud/"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-outline-secondary btn-sm me-2"
          >
            Open Grafana
          </a>
          <button @click="handleLogout" class="btn btn-outline-danger btn-sm">Logout</button>
        </div>
      </div>
    </nav>
  </header>

  <RouterView />
</template>


<style scoped>
</style>
