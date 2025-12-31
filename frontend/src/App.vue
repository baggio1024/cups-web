<template>
  <div class="flex items-center">
    <h1>CUPS 打印</h1>
    <button v-if="view === 'PrintView'" class="btn btn-sm btn-outline ml-4" @click="logout">登出</button>
    <component :is="view" @login-success="onLogin" @logout="onLogout" />
  </div>
</template>

<script>
import LoginView from './views/LoginView.vue'
import PrintView from './views/PrintView.vue'

export default {
  data() {
    return { view: 'LoginView' }
  },
  async mounted() {
    // check existing session on page load; if session present, go straight to PrintView
    try {
      const resp = await fetch('/api/session', { credentials: 'include' })
      if (resp.ok) {
        this.view = 'PrintView'
      }
    } catch (e) {
      // ignore network errors
    }
  },
  components: { LoginView, PrintView },
  methods: {
    onLogin() {
      this.view = 'PrintView'
    },
    onLogout() {
      this.view = 'LoginView'
    },
    async logout() {
      try {
        await fetch('/api/logout', { method: 'POST', credentials: 'include' })
      } catch (e) {
        // ignore errors
      }
      this.onLogout()
    }
  }
}
</script>
