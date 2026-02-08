<template>
  <div class="p-6 space-y-6">
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h2 class="card-title">用户管理</h2>
        <form class="grid grid-cols-1 md:grid-cols-4 gap-3" @submit.prevent="saveUser">
          <input class="input input-bordered" v-model="form.username" :disabled="form.protected" placeholder="登录名" />
          <input class="input input-bordered" type="password" v-model="form.password" :placeholder="isEditing ? '留空不修改密码' : '密码'" />
          <select class="select select-bordered" v-model="form.role" :disabled="form.protected">
            <option value="user">普通用户</option>
            <option value="admin">管理员</option>
          </select>
          <input class="input input-bordered" v-model="form.contactName" placeholder="联系人" />
          <input class="input input-bordered" v-model="form.phone" placeholder="联系电话" />
          <input class="input input-bordered" v-model="form.email" placeholder="邮箱" />
          <input class="input input-bordered hidden" type="number" step="0.01" v-model="form.dailyTopup" placeholder="每日自动充值" />
          <input class="input input-bordered hidden" type="number" step="0.01" v-model="form.monthlyTopup" placeholder="每月自动充值" />
          <input class="input input-bordered hidden" type="number" step="0.01" v-model="form.yearlyTopup" placeholder="每年自动充值" />
          <input class="input input-bordered hidden" type="number" step="0.01" v-model="form.monthlyLimit" placeholder="月度最高消耗" />
          <input class="input input-bordered hidden" type="number" step="0.01" v-model="form.yearlyLimit" placeholder="年度最高消耗" />
          <input class="input input-bordered hidden" type="number" step="0.01" v-model="form.balance" :disabled="isEditing" placeholder="初始余额" />
          <div class="flex gap-2">
            <button class="btn btn-primary" type="submit">{{ isEditing ? '保存' : '新增用户' }}</button>
            <button class="btn btn-ghost" type="button" @click="resetForm">重置</button>
          </div>
        </form>
      </div>

      <div class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>ID</th>
              <th>登录名</th>
              <th>角色</th>
              <th>联系人</th>
              <th>电话</th>
              <th>邮箱</th>
              <th class="hidden">余额</th>
              <th class="hidden">自动充值</th>
              <th class="hidden">限额</th>
              <th>操作</th>
              <th class="hidden">手动充值</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>{{ u.id }}</td>
              <td>{{ u.username }}</td>
              <td>{{ u.role }}</td>
              <td>{{ u.contactName || '-' }}</td>
              <td>{{ u.phone || '-' }}</td>
              <td>{{ u.email || '-' }}</td>
              <td class="hidden">{{ formatCents(u.balanceCents) }}</td>
              <td class="hidden">
                日 {{ formatCents(u.dailyTopupCents) }} /
                月 {{ formatCents(u.monthlyTopupCents) }} /
                年 {{ formatCents(u.yearlyTopupCents) }}
              </td>
              <td class="hidden">
                月 {{ u.monthlyLimitCents ? formatCents(u.monthlyLimitCents) : '未设置' }} /
                年 {{ u.yearlyLimitCents ? formatCents(u.yearlyLimitCents) : '未设置' }}
              </td>
              <td class="space-x-2">
                <button class="btn btn-xs btn-ghost" @click="editUser(u)">编辑</button>
                <button class="btn btn-xs btn-outline btn-error" :disabled="u.username === 'admin'" @click="deleteUser(u)">删除</button>
              </td>
              <td class="hidden space-x-2">
                <input class="input input-bordered input-xs w-24" type="number" step="0.01" v-model="topupAmounts[u.id]" placeholder="金额" />
                <button class="btn btn-xs btn-primary" @click="topupUser(u)">充值</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h2 class="card-title">打印记录</h2>
        <div class="flex flex-wrap gap-3 items-end">
          <input class="input input-bordered" v-model="printFilters.username" placeholder="用户名" />
          <input class="input input-bordered" type="date" v-model="printFilters.start" />
          <input class="input input-bordered" type="date" v-model="printFilters.end" />
          <button class="btn btn-outline" @click="loadPrintRecords">查询</button>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>时间</th>
              <th>用户</th>
              <th>文件</th>
            <th>文件页数</th>
            <th>颜色</th>
            <th>单面/双面</th>
            <th>打印份数</th>
            <th>打印页码</th>
              <th>状态</th>
              <th>下载</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rec in printRecords" :key="rec.id">
              <td>{{ formatDate(rec.createdAt) }}</td>
              <td>{{ rec.username }}</td>
              <td>{{ rec.filename }}</td>
              <td>{{ rec.pages }}</td>
              <td>{{ rec.isColor ? '彩色' : '黑白' }}</td>
            <td>{{ rec.duplex }}</td>
            <td>{{ rec.copies }}</td>
            <td>{{ rec.pageRange }}</td>
              <td>{{ rec.status }}</td>
              <td>
                <a class="link" :href="`/api/print-records/${rec.id}/file`" target="_blank">下载</a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card bg-base-100 shadow hidden">
      <div class="card-body">
        <h2 class="card-title">充值记录</h2>
        <div class="flex flex-wrap gap-3 items-end">
          <input class="input input-bordered" v-model="topupFilters.username" placeholder="用户名" />
          <input class="input input-bordered" type="date" v-model="topupFilters.start" />
          <input class="input input-bordered" type="date" v-model="topupFilters.end" />
          <button class="btn btn-outline" @click="loadTopups">查询</button>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra">
          <thead>
            <tr>
              <th>时间</th>
              <th>用户</th>
              <th>金额</th>
              <th>余额前</th>
              <th>余额后</th>
              <th>类型</th>
              <th>操作人</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rec in topupRecords" :key="rec.id">
              <td>{{ rec.createdAt }}</td>
              <td>{{ rec.username }}</td>
              <td>{{ formatCents(rec.amountCents) }}</td>
              <td>{{ formatCents(rec.balanceBeforeCents) }}</td>
              <td>{{ formatCents(rec.balanceAfterCents) }}</td>
              <td>{{ rec.type }}</td>
              <td>{{ rec.operatorName || 'system' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h2 class="card-title">系统设置</h2>
        <div class="grid grid-cols-1 md:grid-cols-4 gap-3 items-end">
          <label class="form-control hidden">
            <div class="label">
              <span class="label-text">黑白打印单页价格（元/页）</span>
            </div>
            <input class="input input-bordered" type="number" step="0.01" v-model="settings.perPage" placeholder="例如 0.10" />
          </label>
          <label class="form-control hidden">
            <div class="label">
              <span class="label-text">彩色打印单页价格（元/页）</span>
            </div>
            <input class="input input-bordered" type="number" step="0.01" v-model="settings.colorPage" placeholder="例如 0.30" />
          </label>
          <label class="form-control">
            <div class="label">
              <span class="label-text">自动清理天数</span>
            </div>
            <input class="input input-bordered" type="number" step="1" v-model="settings.retentionDays" placeholder="例如 30" />
          </label>
          <div class="flex items-end">
            <button class="btn btn-primary" @click="saveSettings">保存设置</button>
          </div>
        </div>
        <div class="text-sm text-muted mt-2">自动清理会删除打印记录与文件，并压缩数据库。</div>
        <div class="text-sm text-muted mt-1">SESSION_HASH_KEY / SESSION_BLOCK_KEY 通过环境变量配置，未设置会自动生成（仅建议测试环境）。</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      users: [],
      form: {
        id: null,
        username: '',
        password: '',
        role: 'user',
        protected: false,
        contactName: '',
        phone: '',
        email: '',
        balance: '',
        dailyTopup: '',
        monthlyTopup: '',
        yearlyTopup: '',
        monthlyLimit: '',
        yearlyLimit: ''
      },
      topupAmounts: {},
      printFilters: { username: '', start: '', end: '' },
      topupFilters: { username: '', start: '', end: '' },
      printRecords: [],
      topupRecords: [],
      settings: { perPage: '', colorPage: '', retentionDays: '' }
    }
  },
  computed: {
    isEditing() {
      return !!this.form.id
    }
  },
  async mounted() {
    await this.loadUsers()
    await this.loadPrintRecords()
    await this.loadTopups()
    await this.loadSettings()
  },
  methods: {
    getCSRF() {
      const m = document.cookie.match('(^|;)\\s*csrf_token\\s*=\\s*([^;]+)')
      return m ? m.pop() : ''
    },
    formatCents(value) {
      const cents = Number.isFinite(value) ? value : 0
      return (cents / 100).toFixed(2)
    },
    formatDate(dateString) {
      if (!dateString) return '';
      // 将UTC时间转换为本地时间显示
      const date = new Date(dateString);
      return date.toLocaleString();
    },
    toCents(value) {
      const num = parseFloat(String(value).replace(',', '.'))
      if (Number.isNaN(num)) return 0
      return Math.round(num * 100)
    },
    async readError(resp) {
      try {
        const data = await resp.json()
        return data.error || resp.statusText
      } catch (e) {
        try {
          const text = await resp.text()
          return text || resp.statusText
        } catch (err) {
          return resp.statusText
        }
      }
    },
    resetForm() {
      this.form = {
        id: null,
        username: '',
        password: '',
        role: 'user',
        protected: false,
        contactName: '',
        phone: '',
        email: '',
        balance: '',
        dailyTopup: '',
        monthlyTopup: '',
        yearlyTopup: '',
        monthlyLimit: '',
        yearlyLimit: ''
      }
    },
    editUser(user) {
      this.form = {
        id: user.id,
        username: user.username,
        password: '',
        role: user.role,
        protected: user.username === 'admin',
        contactName: user.contactName || '',
        phone: user.phone || '',
        email: user.email || '',
        balance: '',
        dailyTopup: this.formatCents(user.dailyTopupCents),
        monthlyTopup: this.formatCents(user.monthlyTopupCents),
        yearlyTopup: this.formatCents(user.yearlyTopupCents),
        monthlyLimit: user.monthlyLimitCents ? this.formatCents(user.monthlyLimitCents) : '',
        yearlyLimit: user.yearlyLimitCents ? this.formatCents(user.yearlyLimitCents) : ''
      }
    },
    async loadUsers() {
      const resp = await fetch('/api/admin/users', { credentials: 'include' })
      if (!resp.ok) {
        if (resp.status === 401) this.$emit('logout')
        return
      }
      this.users = await resp.json()
    },
    async saveUser() {
      const payload = {
        username: this.form.username,
        password: this.form.password,
        role: this.form.role,
        contactName: this.form.contactName,
        phone: this.form.phone,
        email: this.form.email,
        balanceCents: this.toCents(this.form.balance),
        dailyTopupCents: this.toCents(this.form.dailyTopup),
        monthlyTopupCents: this.toCents(this.form.monthlyTopup),
        yearlyTopupCents: this.toCents(this.form.yearlyTopup),
        monthlyLimitCents: this.toCents(this.form.monthlyLimit),
        yearlyLimitCents: this.toCents(this.form.yearlyLimit)
      }
      const isEditing = this.isEditing
      const url = isEditing ? `/api/admin/users/${this.form.id}` : '/api/admin/users'
      const method = isEditing ? 'PUT' : 'POST'
      const resp = await fetch(url, {
        method,
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.getCSRF()
        },
        body: JSON.stringify(payload)
      })
      if (!resp.ok) {
        const msg = await this.readError(resp)
        alert(msg)
        if (resp.status === 401) this.$emit('logout')
        return
      }
      await this.loadUsers()
      this.resetForm()
    },
    async deleteUser(user) {
      if (!confirm(`确认删除用户 ${user.username} ?`)) return
      const resp = await fetch(`/api/admin/users/${user.id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: { 'X-CSRF-Token': this.getCSRF() }
      })
      if (!resp.ok) {
        const msg = await this.readError(resp)
        alert(msg)
        if (resp.status === 401) this.$emit('logout')
        return
      }
      await this.loadUsers()
    },
    async topupUser(user) {
      const amount = this.toCents(this.topupAmounts[user.id])
      if (amount <= 0) {
        alert('请输入正确的充值金额')
        return
      }
      const resp = await fetch(`/api/admin/users/${user.id}/topup`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.getCSRF()
        },
        body: JSON.stringify({ amountCents: amount })
      })
      if (!resp.ok) {
        const msg = await this.readError(resp)
        alert(msg)
        if (resp.status === 401) this.$emit('logout')
        return
      }
      this.topupAmounts[user.id] = ''
      await this.loadUsers()
      await this.loadTopups()
    },
    async loadPrintRecords() {
      const params = new URLSearchParams()
      if (this.printFilters.username) params.set('username', this.printFilters.username)
      if (this.printFilters.start) params.set('start', this.printFilters.start)
      if (this.printFilters.end) params.set('end', this.printFilters.end)
      const resp = await fetch(`/api/admin/print-records?${params.toString()}`, { credentials: 'include' })
      if (!resp.ok) {
        if (resp.status === 401) this.$emit('logout')
        return
      }
      this.printRecords = await resp.json()
    },
    async loadTopups() {
      const params = new URLSearchParams()
      if (this.topupFilters.username) params.set('username', this.topupFilters.username)
      if (this.topupFilters.start) params.set('start', this.topupFilters.start)
      if (this.topupFilters.end) params.set('end', this.topupFilters.end)
      const resp = await fetch(`/api/admin/topups?${params.toString()}`, { credentials: 'include' })
      if (!resp.ok) {
        if (resp.status === 401) this.$emit('logout')
        return
      }
      this.topupRecords = await resp.json()
    },
    async loadSettings() {
      const resp = await fetch('/api/admin/settings', { credentials: 'include' })
      if (!resp.ok) {
        if (resp.status === 401) this.$emit('logout')
        return
      }
      const data = await resp.json()
      this.settings.perPage = this.formatCents(data.perPageCents || 0)
      this.settings.colorPage = this.formatCents(data.colorPageCents || 0)
      this.settings.retentionDays = String(data.retentionDays || 0)
    },
    async saveSettings() {
      const payload = {
        perPageCents: this.toCents(this.settings.perPage),
        colorPageCents: this.toCents(this.settings.colorPage),
        retentionDays: parseInt(this.settings.retentionDays || '0', 10)
      }
      const resp = await fetch('/api/admin/settings', {
        method: 'PUT',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': this.getCSRF()
        },
        body: JSON.stringify(payload)
      })
      if (!resp.ok) {
        const msg = await this.readError(resp)
        alert(msg)
        if (resp.status === 401) this.$emit('logout')
        return
      }
      await this.loadSettings()
    }
  }
}
</script>
