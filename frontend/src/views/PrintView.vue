<template>
  <div class="p-6">
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h2 class="card-title">Print</h2>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-3 mb-4">
          <div class="alert bg-base-200">
            <div class="space-y-1 text-sm">
              <div>余额: {{ formatCents(balanceCents) }}</div>
              <div>单页价格: {{ formatCents(perPageCents) }}</div>
              <div>本月已用: {{ formatCents(monthSpentCents) }}</div>
              <div>本年已用: {{ formatCents(yearSpentCents) }}</div>
            </div>
          </div>
          <div class="alert bg-base-200">
            <div class="space-y-1 text-sm">
              <div>月度限额: {{ monthlyLimitCents ? formatCents(monthlyLimitCents) : '未设置' }}</div>
              <div>年度限额: {{ yearlyLimitCents ? formatCents(yearlyLimitCents) : '未设置' }}</div>
            </div>
          </div>
          <div class="alert bg-base-200">
            <div class="text-sm" v-if="estimating">估算中…</div>
            <div class="space-y-1 text-sm" v-else-if="estimate">
              <div>预估页数: {{ estimate.pages }} <span v-if="estimate.estimated">(估算)</span></div>
              <div>预估费用: {{ formatCents(estimate.costCents) }}</div>
              <div v-if="estimate.insufficientBalance" class="text-error">余额不足</div>
              <div v-if="estimate.wouldExceedMonthly" class="text-warning">超过月度限额</div>
              <div v-if="estimate.wouldExceedYearly" class="text-warning">超过年度限额</div>
            </div>
            <div class="text-sm" v-else>选择文件后显示估算</div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="col-span-1 space-y-4">
            <label class="label">
              <span class="label-text">打印机</span>
            </label>
            <select v-model="printer" class="select select-bordered w-full">
              <option v-for="p in printers" :key="p.uri" :value="p.uri">{{ p.name }} — {{ p.uri }}</option>
            </select>

            <div>
              <label class="label">
                <span class="label-text">文件</span>
              </label>
              <input type="file" ref="file" @change="onFileChange" class="file-input file-input-bordered w-full" />
            </div>

            <div class="space-x-2">
              <button class="btn btn-primary" :disabled="!canPrint || converting" @click="uploadAndPrint">打印</button>
              <button class="btn" :disabled="!canConvert" @click="convertToPdf">转换</button>
              <a v-if="previewUrl" :href="previewUrl" :download="downloadName" class="btn btn-ghost">下载预览</a>
            </div>

            <div class="text-sm text-muted">{{ msg }}</div>

            <div class="mt-4">
              <label class="label"><span class="label-text">转换状态</span></label>
              <div v-if="converting" class="alert alert-info">转换中…</div>
              <div v-if="converted" class="alert alert-success">已转换为 PDF</div>
            </div>
          </div>

          <div class="col-span-2">
            <label class="label"><span class="label-text">Preview</span></label>
            <div class="preview-container p-2">
              <div v-if="previewType === 'image'" class="flex items-center justify-center">
                <img :src="previewUrl" alt="preview" class="max-h-[600px] max-w-full" />
              </div>
              <div v-else-if="previewType === 'pdf'">
                <iframe :src="previewUrl" style="width:100%; height:600px;" frameborder="0"></iframe>
              </div>
              <div v-else-if="previewType === 'text'" class="p-4 whitespace-pre-wrap overflow-auto h-64">
                {{ textPreview }}
              </div>
              <div v-else class="p-4 text-muted">无预览可用</div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script>
import { jsPDF } from 'jspdf'

export default {
  data() {
    return {
      printer: '',
      printers: [],
      msg: '',
      selectedFile: null,
      previewUrl: '',
      previewType: '', // 'pdf' | 'image' | 'text' | ''
      textPreview: '',
      converting: false,
      converted: false,
      pdfBlob: null,
      downloadName: '',
      balanceCents: 0,
      perPageCents: 0,
      monthSpentCents: 0,
      yearSpentCents: 0,
      monthlyLimitCents: 0,
      yearlyLimitCents: 0,
      estimate: null,
      estimating: false
    }
  },
  computed: {
    canPrint() {
      const hasFile = !!this.printer && (!!this.pdfBlob || !!this.selectedFile)
      if (!hasFile) return false
      if (this.estimating) return false
      if (this.estimate) {
        if (this.estimate.insufficientBalance) return false
        if (this.estimate.wouldExceedMonthly || this.estimate.wouldExceedYearly) return false
      }
      return true
    },
    canConvert() {
      // disable convert when no file, while converting, or if file is already PDF
      return !!this.selectedFile && !this.converting && this.selectedFile.type !== 'application/pdf'
    }
  },
  async mounted() {
    await this.loadProfile()
    try {
      const resp = await fetch('/api/printers', { credentials: 'include' })
      if (resp.ok) {
        this.printers = await resp.json()
        const last = localStorage.getItem('last_printer')
        if (last) this.printer = last
        else if (this.printers.length > 0) this.printer = this.printers[0].uri
      } else if (resp.status === 401) {
        // session expired / not logged in; notify parent to switch to login view
        this.$emit('logout')
      } else {
        this.msg = '加载打印机失败'
      }
    } catch (e) {
      this.msg = '加载打印机失败: ' + e.message
    }
  },
  methods: {
    async loadProfile() {
      try {
        const resp = await fetch('/api/me', { credentials: 'include' })
        if (!resp.ok) {
          if (resp.status === 401) this.$emit('logout')
          return
        }
        const data = await resp.json()
        this.balanceCents = data.balanceCents || 0
        this.perPageCents = data.perPageCents || 0
        this.monthSpentCents = data.monthSpentCents || 0
        this.yearSpentCents = data.yearSpentCents || 0
        this.monthlyLimitCents = data.monthlyLimitCents || 0
        this.yearlyLimitCents = data.yearlyLimitCents || 0
      } catch (e) {
        // ignore
      }
    },
    getCSRF() {
      const m = document.cookie.match('(^|;)\\s*csrf_token\\s*=\\s*([^;]+)')
      return m ? m.pop() : ''
    },
    formatCents(value) {
      const cents = Number.isFinite(value) ? value : 0
      return (cents / 100).toFixed(2)
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
    clearPreview() {
      // revoke any existing object URL and reset preview-related state
      if (this.previewUrl) {
        try {
          URL.revokeObjectURL(this.previewUrl)
        } catch (e) {
          // ignore
        }
      }
      this.previewUrl = ''
      this.previewType = ''
      this.textPreview = ''
      this.pdfBlob = null
      this.converted = false
      this.selectedFile = null
      this.downloadName = ''
      this.estimate = null
      this.estimating = false
    },
    onFileChange(e) {
      const f = e.target.files[0]
      this.clearPreview()
      if (!f) return
      this.selectedFile = f
      this.downloadName = f.name.replace(/\.[^/.]+$/, '') + '.pdf'

      if (f.type === 'application/pdf') {
        this.previewUrl = URL.createObjectURL(f)
        this.previewType = 'pdf'
        this.pdfBlob = f
        this.converted = true
      } else if (f.type.startsWith('image/')) {
        this.previewUrl = URL.createObjectURL(f)
        this.previewType = 'image'
        this.pdfBlob = null
        this.converted = false
      } else if (this.isOfficeFile(f)) {
        // Office files: preview not available in-browser; show a notice
        this.previewType = 'text'
        this.textPreview = 'Office 文档（无法预览）。点击“转换”生成 PDF。'
        this.pdfBlob = null
        this.converted = false
      } else if (f.type.startsWith('text/') || /\.(txt|md|html)$/i.test(f.name)) {
        const reader = new FileReader()
        reader.onload = () => {
          this.textPreview = reader.result
          this.previewType = 'text'
        }
        reader.readAsText(f)
        this.pdfBlob = null
        this.converted = false
      } else {
        // fallback attempt to read as text for preview
        const reader = new FileReader()
        reader.onload = () => {
          const text = typeof reader.result === 'string' ? reader.result : ''
          this.textPreview = text.slice(0, 10000) || 'No preview available'
          this.previewType = 'text'
        }
        reader.readAsText(f)
        this.pdfBlob = null
        this.converted = false
      }
      this.estimatePrice()
    },
    async estimatePrice() {
      const fileForEstimate = this.pdfBlob || this.selectedFile
      if (!fileForEstimate) return
      this.estimating = true
      const form = new FormData()
      const name = this.downloadName || fileForEstimate.name || 'document.pdf'
      form.append('file', fileForEstimate, name)
      try {
        const resp = await fetch('/api/estimate', {
          method: 'POST',
          body: form,
          credentials: 'include',
          headers: { 'X-CSRF-Token': this.getCSRF() }
        })
        if (!resp.ok) {
          this.msg = await this.readError(resp)
          if (resp.status === 401) this.$emit('logout')
          return
        }
        const data = await resp.json()
        this.estimate = data
        this.perPageCents = data.perPageCents || this.perPageCents
        this.balanceCents = data.balanceCents || this.balanceCents
        this.monthSpentCents = data.monthSpentCents || this.monthSpentCents
        this.yearSpentCents = data.yearSpentCents || this.yearSpentCents
        this.monthlyLimitCents = data.monthlyLimitCents || this.monthlyLimitCents
        this.yearlyLimitCents = data.yearlyLimitCents || this.yearlyLimitCents
      } catch (e) {
        this.msg = '估算失败: ' + e.message
      } finally {
        this.estimating = false
      }
    },
    async convertToPdf() {
      if (!this.selectedFile) { this.msg = 'No file selected'; return }
      this.converting = true
      this.msg = ''
      try {
        const f = this.selectedFile
        let blob = null

        // Office file types will be converted on the server
        if (this.isOfficeFile(f)) {
          blob = await this.convertOfficeToPdf(f)
        } else if (f.type.startsWith('image/')) {
          blob = await this.imageFileToPdfBlob(f)
        } else if (f.type.startsWith('text/') || /\.(txt|md|html)$/i.test(f.name)) {
          const text = await f.text()
          blob = this.textToPdfBlob(text)
        } else {
          // general fallback: attempt to read as text and convert
          try {
            const text = await f.text()
            blob = this.textToPdfBlob(text)
          } catch (e) {
            throw new Error('Unsupported file type for conversion')
          }
        }

        this.pdfBlob = blob
        this.previewUrl = URL.createObjectURL(blob)
        this.previewType = 'pdf'
        this.converted = true
        this.msg = '已准备好转换'
        this.estimatePrice()
      } catch (e) {
        this.msg = '转换失败: ' + e.message
      } finally {
        this.converting = false
      }
    },
    isOfficeFile(f) {
      return /\.(docx?|pptx?|xlsx?)$/i.test(f.name) || [
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'application/msword',
        'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        'application/vnd.ms-powerpoint',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'application/vnd.ms-excel'
      ].includes(f.type)
    },
    async convertOfficeToPdf(file) {
      const fd = new FormData()
      fd.append('file', file, file.name)
      try {
        const resp = await fetch('/api/convert', {
          method: 'POST',
          body: fd,
          credentials: 'include',
          headers: { 'X-CSRF-Token': this.getCSRF() }
        })
        if (!resp.ok) {
          const t = await resp.text()
          throw new Error('server conversion failed: ' + t)
        }
        const blob = await resp.blob()
        if (blob.type !== 'application/pdf') {
          // sometimes servers may not set mime; still accept
          // but warn in msg
          this.msg = '已收到已转换文件 (mime: ' + blob.type + ')'
        }
        return blob
      } catch (e) {
        throw e
      }
    },
    imageFileToPdfBlob(file) {
      return new Promise((resolve, reject) => {
        const img = new Image()
        img.onload = () => {
          // create canvas sized to image
          const canvas = document.createElement('canvas')
          canvas.width = img.width
          canvas.height = img.height
          const ctx = canvas.getContext('2d')
          ctx.drawImage(img, 0, 0)
          const imgData = canvas.toDataURL('image/jpeg', 1.0)
          // jsPDF works in "pt" units by default; we'll use px units and convert via options
          const doc = new jsPDF({ unit: 'px', format: [img.width, img.height] })
          doc.addImage(imgData, 'JPEG', 0, 0, img.width, img.height)
          const blob = doc.output('blob')
          resolve(blob)
        }
        img.onerror = () => reject(new Error('Failed to load image for conversion'))
        img.src = URL.createObjectURL(file)
      })
    },
    textToPdfBlob(text) {
      const doc = new jsPDF()
      const lines = doc.splitTextToSize(text || '', 180)
      doc.text(lines, 10, 10)
      return doc.output('blob')
    },
    async uploadAndPrint() {
      if (!this.printer) { this.msg = '请选择打印机'; return }
      let fileToSend = null
      let filename = ''
      if (this.pdfBlob) {
        fileToSend = this.pdfBlob
        filename = this.downloadName || (this.selectedFile && this.selectedFile.name.replace(/\.[^/.]+$/, '') + '.pdf') || 'document.pdf'
      } else if (this.selectedFile) {
        // if PDF was not created but original file is PDF (maybe set earlier)
        fileToSend = this.selectedFile
        filename = this.selectedFile.name
      } else {
        this.msg = '没有可打印的文件'
        return
      }

      const form = new FormData()
      form.append('file', fileToSend, filename)
      form.append('printer', this.printer)

      try {
        const resp = await fetch('/api/print', {
          method: 'POST',
          body: form,
          credentials: 'include',
          headers: { 'X-CSRF-Token': this.getCSRF() }
        })
        if (!resp.ok) {
          this.msg = await this.readError(resp)
          if (resp.status === 401) this.$emit('logout')
          return
        }
        const j = await resp.json()
        this.msg = '任务已加入队列: ' + (j.jobId || '')
        this.balanceCents = j.balanceCents || this.balanceCents
        this.monthSpentCents = j.monthSpentCents || this.monthSpentCents
        this.yearSpentCents = j.yearSpentCents || this.yearSpentCents
        this.estimate = null
        localStorage.setItem('last_printer', this.printer)
      } catch (e) {
        this.msg = e.message
      }
    }
  }
}
</script>
