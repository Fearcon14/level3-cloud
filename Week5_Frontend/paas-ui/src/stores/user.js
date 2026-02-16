import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', () => {
  const savedUser = localStorage.getItem('paas-user')
  const username = ref(savedUser || null)


  function login(name) {
    username.value = name
    localStorage.setItem('paas-user', name)
  }

  function logout() {
    username.value = null
    localStorage.removeItem('paas-user')
  }

  return { username, login, logout }
})
