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
        <span class="navbar-brand">PaaS Dashboard</span>
        <div class="d-flex align-items-center">
          <span class="me-3">Hello, <strong>{{ userStore.username }}</strong></span>
          <button @click="handleLogout" class="btn btn-outline-danger btn-sm">Logout</button>
        </div>
      </div>
    </nav>
  </header>

  <RouterView />
</template>

<style scoped>
</style>
