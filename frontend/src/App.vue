<template>
  <div>
    <h1>CUPS Web</h1>
    <component :is="view" @login-success="onLogin" />
  </div>
</template>

<script>
import LoginView from './views/LoginView.vue'
import PrintView from './views/PrintView.vue'

export default {
  data() {
    return { view: 'LoginView' }
  },
  components: { LoginView, PrintView },
  methods: {
    onLogin() {
      this.view = 'PrintView'
    }
  }
}
</script>
