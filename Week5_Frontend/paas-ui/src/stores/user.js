import { ref } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'

export const useUserStore = defineStore('user', () => {
  const savedUser = localStorage.getItem('paas-user')
  const savedToken = localStorage.getItem('paas-token')

  const username = ref(savedUser || null)
  const token = ref(savedToken || null)

  // Initialize axios header if token exists
  if (token.value) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  async function login(name, password) {
    try {
      const response = await axios.post('/api/login', {
        username: name,
        password: password
      })

      const authToken = response.data.token

      username.value = name
      token.value = authToken

      localStorage.setItem('paas-user', name)
      localStorage.setItem('paas-token', authToken)

      axios.defaults.headers.common['Authorization'] = `Bearer ${authToken}`
      return true
    } catch (error) {
      console.error('Login failed:', error)
      throw error // Re-throw to handle in component
    }
  }

  function logout() {
    username.value = null
    token.value = null
    localStorage.removeItem('paas-user')
    localStorage.removeItem('paas-token')
    delete axios.defaults.headers.common['Authorization']
  }

  return { username, token, login, logout }
})
