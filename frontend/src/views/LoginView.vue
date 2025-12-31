<template>
  <div class="flex items-center justify-center min-h-screen bg-base-200 p-6">
    <div class="card w-full max-w-md bg-base-100 shadow-md">
      <div class="card-body">
        <h2 class="card-title">Login</h2>
        <form @submit.prevent="login" class="space-y-4">
          <div>
            <label class="label">
              <span class="label-text">Username</span>
            </label>
            <input class="input input-bordered w-full" v-model="username" />
          </div>
          <div>
            <label class="label">
              <span class="label-text">Password</span>
            </label>
            <input type="password" class="input input-bordered w-full" v-model="password" />
          </div>
          <div class="flex items-center justify-between">
            <button class="btn btn-primary" type="submit">Login</button>
            <div v-if="error" class="text-error">{{ error }}</div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() { return { username: '', password: '', error: '' } },
  methods: {
    async login() {
      this.error = ''
      try {
        const resp = await fetch('/api/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: this.username, password: this.password }),
          credentials: 'include'
        })
        if (!resp.ok) {
          this.error = 'Login failed'
          return
        }
        this.$emit('login-success')
      } catch (e) {
        this.error = e.message
      }
    }
  }
}
</script>
