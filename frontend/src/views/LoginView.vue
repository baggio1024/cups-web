<template>
  <div class="min-h-screen flex items-center justify-center relative overflow-hidden bg-paper selection:bg-cmyk-c selection:text-white">
    
    <!-- Sophisticated Technical Background -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <!-- Fine Grid -->
      <div class="absolute inset-0 opacity-[0.03] bg-[linear-gradient(to_right,#000_1px,transparent_1px),linear-gradient(to_bottom,#000_1px,transparent_1px)] bg-[size:40px_40px]"></div>
      
      <!-- Radial Gradient Atmosphere -->
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(0,147,213,0.03),transparent_70%)]"></div>

      <!-- Floating Technical Elements -->
      <div class="absolute top-[10%] left-[5%] animate-fade-in opacity-20 hidden lg:block">
        <div class="font-mono text-[10px] space-y-1">
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 bg-cmyk-c"></span>
            <span>CYAN_CHANNEL 001.293</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 bg-cmyk-m"></span>
            <span>MAGENTA_CHANNEL 004.882</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-2.5 h-2.5 bg-cmyk-y"></span>
            <span>YELLOW_CHANNEL 002.551</span>
          </div>
        </div>
      </div>

      <!-- Registration Marks -->
      <div class="absolute top-10 right-10 flex gap-4 opacity-10 font-mono text-xs">
        <div class="flex flex-col items-center">
          <div class="w-8 h-8 rounded-full border border-black flex items-center justify-center relative mb-1">
            <div class="absolute w-full h-[1px] bg-black"></div>
            <div class="absolute h-full w-[1px] bg-black"></div>
          </div>
          <span>MARK_01</span>
        </div>
      </div>

      <div class="absolute bottom-10 left-10 opacity-10 font-mono text-[10px] tracking-tighter">
        <div>COORDINATES: [40.7128° N, 74.0060° W]</div>
        <div>STABILITY: OPTIMAL</div>
      </div>
    </div>

    <!-- Main Container -->
    <main class="relative z-20 w-full max-w-[440px] px-6 py-12">
      <!-- Glass Card -->
      <div class="relative group bg-white/40 backdrop-blur-2xl border border-white/50 shadow-[0_32px_64px_-16px_rgba(0,0,0,0.1)] rounded-[32px] p-8 md:p-12 transition-all duration-700 hover:shadow-[0_48px_80px_-24px_rgba(0,0,0,0.15)] animate-scale-in">
        
        <!-- Subtle Glow Edge -->
        <div class="absolute -inset-px bg-gradient-to-br from-white/80 via-transparent to-black/5 rounded-[32px] pointer-events-none"></div>

        <!-- Branding Header -->
        <header class="mb-12 text-center opacity-0 animate-slide-up">
          <div class="relative inline-flex items-center justify-center mb-8">
            <!-- Animated rings -->
            <div class="absolute inset-0 -m-3 border border-cmyk-c/20 rounded-full animate-[spin_12s_linear_infinite]"></div>
            <div class="absolute inset-0 -m-6 border border-cmyk-m/10 rounded-full animate-[spin_20s_linear_infinite_reverse]"></div>
            
            <div class="relative w-20 h-20 bg-gradient-to-br from-cmyk-c to-cmyk-m p-0.5 rounded-2xl shadow-xl rotate-3 transition-transform hover:rotate-0 duration-500">
              <div class="w-full h-full bg-paper rounded-[14px] flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-cmyk-k">
                  <polyline points="6 9 6 2 18 2 18 9"></polyline>
                  <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"></path>
                  <rect x="6" y="14" width="12" height="8"></rect>
                </svg>
              </div>
            </div>
          </div>

          <h1 class="text-4xl font-extrabold tracking-tight text-cmyk-k mb-3">
            CUPS<span class="text-cmyk-c">.</span>WEB
          </h1>
          <p class="text-[10px] font-bold tracking-[0.3em] text-ink-light uppercase flex items-center justify-center gap-3">
            <span class="w-8 h-[1px] bg-black/5"></span>
            统一打印管理平台
            <span class="w-8 h-[1px] bg-black/5"></span>
          </p>
        </header>

        <!-- Access Form -->
        <form @submit.prevent="login" class="space-y-6 opacity-0 animate-slide-up delay-100">
          <!-- Input Field: Username -->
          <div class="space-y-2 group/input">
            <div class="relative">
              <input 
                v-model="username"
                type="text" 
                class="w-full bg-paper/50 border border-black/5 rounded-2xl px-5 py-4 text-sm focus:bg-white focus:outline-none focus:border-cmyk-c/50 focus:ring-4 focus:ring-cmyk-c/5 transition-all duration-300 placeholder:text-black/20"
                placeholder="请输入用户名"
                :disabled="loading"
                required
              />
              <div class="absolute right-4 top-1/2 -translate-y-1/2 opacity-20 pointer-events-none">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
              </div>
            </div>
          </div>
          
          <!-- Input Field: Password -->
          <div class="space-y-2 group/input">
            <div class="relative">
              <input 
                v-model="password"
                type="password" 
                class="w-full bg-paper/50 border border-black/5 rounded-2xl px-5 py-4 text-sm focus:bg-white focus:outline-none focus:border-cmyk-m/50 focus:ring-4 focus:ring-cmyk-m/5 transition-all duration-300 placeholder:text-black/20"
                placeholder="请输入密码"
                :disabled="loading"
                required
              />
              <div class="absolute right-4 top-1/2 -translate-y-1/2 opacity-20 pointer-events-none">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
              </div>
            </div>
          </div>

          <!-- Submit Button -->
          <div class="pt-4">
            <button 
              type="submit" 
              :disabled="loading"
              class="relative w-full h-14 bg-cmyk-k rounded-2xl overflow-hidden group/btn disabled:opacity-80 transition-all duration-300 active:scale-[0.98]"
            >
              <!-- Hover Gradient Effect -->
              <div class="absolute inset-0 opacity-0 group-hover/btn:opacity-100 bg-gradient-to-r from-cmyk-c via-cmyk-m to-cmyk-y transition-opacity duration-500"></div>
              
              <div class="relative h-full flex items-center justify-center">
                <div v-if="!loading" class="flex items-center gap-3 text-white font-bold tracking-widest text-xs">
                  <span>开启安全会话</span>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="transition-transform group-hover/btn:translate-x-1"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
                </div>
                <div v-else class="flex items-center gap-3">
                  <div class="w-5 h-5 border-2 border-white/20 border-t-white rounded-full animate-spin"></div>
                  <span class="text-white font-bold tracking-[0.2em] text-[10px]">身份验证中...</span>
                </div>
              </div>

              <!-- Button Shine Effect -->
              <div class="absolute top-0 -left-[100%] w-[50%] h-full bg-gradient-to-r from-transparent via-white/10 to-transparent skew-x-[-25deg] group-hover/btn:left-[100%] transition-all duration-700"></div>
            </button>
          </div>

          <!-- Error Message -->
          <div v-if="error" class="bg-red-50/50 backdrop-blur-sm border border-red-100 rounded-xl p-4 animate-shake">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 flex-shrink-0 bg-red-100 text-red-600 rounded-lg flex items-center justify-center">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
              </div>
              <p class="text-[11px] font-bold text-red-600 leading-tight uppercase font-mono">
                System Alert: {{ error }}
              </p>
            </div>
          </div>
        </form>

        <!-- Footer -->
        <footer class="mt-12 text-center opacity-0 animate-slide-up delay-200">
          <div class="flex items-center justify-center gap-2 mb-4">
            <span class="w-1.5 h-1.5 rounded-full bg-cmyk-c"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-cmyk-m"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-cmyk-y"></span>
            <span class="w-1.5 h-1.5 rounded-full bg-cmyk-k"></span>
          </div>
          <p class="text-[9px] font-bold text-ink-light/40 uppercase tracking-widest">
            <a href="https://github.com/baggio1024/cups-web" class="text-[10px] font-bold text-ink-light/40 uppercase tracking-widest hover:text-cmyk-m transition-colors">云打印协议系统 v1.0</a>
          </p>
        </footer>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['login-success'])

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const login = async () => {
  error.value = ''
  loading.value = true
  
  try {
    const resp = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username: username.value, 
        password: password.value 
      }),
      credentials: 'include'
    })

    if (!resp.ok) {
      error.value = '安全验证失败：请检查用户名和密码'
      loading.value = false
      return
    }

    emit('login-success')
  } catch (e) {
    error.value = `系统异常: ${e.message}`
    loading.value = false
  }
}
</script>

<style scoped>
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-4px); }
  75% { transform: translateX(4px); }
}

.animate-shake {
  animation: shake 0.2s cubic-bezier(.36,.07,.19,.97) both;
}

/* Custom cubic-bezier for smoother feel */
.animate-slide-up {
  opacity: 1; /* Override opacity-0 from tailwind once animation finishes */
}
</style>

