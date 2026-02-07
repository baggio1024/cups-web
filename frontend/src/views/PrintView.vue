<template>
  <div class="p-6">
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="mb-2"></div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div class="col-span-1 space-y-3">
            <label class="label">
              <span class="label-text">打印机</span>
            </label>
            <select v-model="printer" class="select select-bordered w-full select-sm" :class="{ 'select-error': !printer || printers.length === 0 }">
              <option value="" disabled>请选择打印机</option>
              <option v-for="p in printers" :key="p.uri" :value="p.uri">{{ p.name }} — {{ p.uri }}</option>
            </select>
            <div>
              <label class="label">
                <span class="label-text">文件</span>
              </label>
              <input type="file" ref="file" @change="onFileChange" class="file-input file-input-bordered w-full file-input-sm" :class="{ 'input-error': !selectedFile && !pdfBlob }" />
            </div>
            <div>
              <label class="label">
                <span class="label-text">打印选项</span>
              </label>
              <select v-model="sides" class="select select-bordered w-full select-sm" :class="{ 'select-error': !sides }">
                <option value="">请选择...</option>
                <option value="one-sided">单面打印</option>
                <option value="two-sided-long-edge">双面打印（长边翻转）</option>
                <option value="two-sided-short-edge">双面打印（短边翻转）</option>
              </select>
            </div>
            <div>
              <label class="label">
                <span class="label-text">打印份数</span>
              </label>
              <input type="number" v-model.number="copies" min="1" max="100" class="input input-bordered w-full input-sm" :class="{ 'input-error': !copies || copies < 1 }" />
            </div>

            <div>
              <label class="label">
                <span class="label-text">打印页面</span>
              </label>
              <select v-model="pageRange" class="select select-bordered w-full select-sm" :class="{ 'select-error': !pageRange || (pageRange !== 'all' && pageRange !== 'custom') }">
                <option value="all">全部页面</option>
                <option value="custom">自定义页面范围</option>
              </select>
              <input v-if="pageRange === 'custom'" type="text" v-model="customPageRange" placeholder="例如: 1-3,5,7-9" class="input input-bordered w-full mt-1 input-sm" :class="{ 'input-error': pageRange === 'custom' && !customPageRange }" />
              <div v-if="totalPages > 0" class="text-sm mt-1 p-2 bg-base-200 rounded">
                <strong>文件总页数: {{ totalPages }} 页</strong>
              </div>
              <div v-else-if="calculatingPages" class="text-sm mt-1 p-2 bg-info/10 rounded">
                正在计算页数...
              </div>
            </div>

            

            <div>
              <label class="label">
                <span class="label-text">颜色模式</span>
              </label>
              <select v-model="isColor" class="select select-bordered w-full select-sm">
                <option :value="false">黑白打印</option>
                <option :value="true">彩色打印</option>
              </select>
            </div>

            <div class="space-x-2 mt-2">
              <button class="btn btn-primary btn-sm" :disabled="!canPrint || converting" @click="uploadAndPrint">打印</button>
              <button class="btn btn-sm" :disabled="!canConvert" @click="convertToPdf">转换</button>
              <a v-if="previewUrl" :href="previewUrl" :download="downloadName" class="btn btn-ghost btn-sm">下载预览</a>
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
                <iframe :src="previewUrl" style="width:100%; height:500px;" frameborder="0"></iframe>
              </div>
              <div v-else-if="previewType === 'text'" class="p-4 whitespace-pre-wrap overflow-auto h-48">
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
      colorPageCents: 0,
      monthSpentCents: 0,
      yearSpentCents: 0,
      monthlyLimitCents: 0,
      yearlyLimitCents: 0,
      estimate: null,
      estimating: false,
      sides: '',
      isColor: false,
      copies: 1,
      pageRange: 'all',
      customPageRange: '',
      totalPages: 0,
      calculatingPages: false
    }
  },
  computed: {
    canPrint() {
      if (!this.printer) return false
      const hasFile = !!this.pdfBlob || !!this.selectedFile
      if (!hasFile) return false
      if (!this.sides) return false
      if (!this.copies || this.copies < 1) return false
      if (!this.pageRange || (this.pageRange !== 'all' && this.pageRange !== 'custom')) return false
      if (this.pageRange === 'custom' && !this.customPageRange) return false
      return true
    },
    canConvert() {
      // disable convert when no file, while converting, or if file is already PDF
      return !!this.selectedFile && !this.converting && this.selectedFile.type !== 'application/pdf'
    }
  },
  async mounted() {
    // 预加载PDF.js库
    try {
      const script = document.createElement('script')
      script.src = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.min.js'
      script.onload = () => {
        // 设置worker
        window.pdfjsLib = window.pdfjsLib || window.pdf
        if (window.pdfjsLib) {
          window.pdfjsLib.GlobalWorkerOptions = window.pdfjsLib.GlobalWorkerOptions || {}
          window.pdfjsLib.GlobalWorkerOptions.workerSrc = 'https://cdnjs.cloudflare.com/ajax/libs/pdf.js/3.11.174/pdf.worker.min.js'
        }
      }
      document.head.appendChild(script)
    } catch (error) {
      console.error('Error loading PDF.js:', error)
    }
    
    await this.loadProfile()
    try {
      const resp = await fetch('/api/printers', { credentials: 'include' })
      if (resp.ok) {
        this.printers = await resp.json()
        if (this.printers.length > 0) this.printer = this.printers[0].uri
        else this.printer = ''
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
        this.colorPageCents = data.colorPageCents || 0
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
      this.totalPages = 0
      this.calculatingPages = false
    },
    onFileChange(e) {
      const f = e.target.files[0]
      if (!f) return
      
      // 清除之前的预览，但保留页数直到新文件处理完成
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
      this.selectedFile = f
      this.downloadName = f.name.replace(/\.[^/.]+$/, '') + '.pdf'
      this.estimate = null
      this.estimating = false
      this.totalPages = 0 // 重置页数，稍后会更新

      if (f.type === 'application/pdf') {
        this.previewUrl = URL.createObjectURL(f)
        this.previewType = 'pdf'
        this.pdfBlob = f
        this.converted = true
        // 获取PDF页数并立即更新UI
        this.getPdfPageCount(f).then(count => {
          this.totalPages = count
          // 强制Vue更新UI
          this.$forceUpdate()
        })
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
      // removed estimation call
    },
    // estimation removed
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
        
        // 获取转换后的PDF页数并立即更新UI
        this.getPdfPageCount(blob).then(count => {
          this.totalPages = count
          // 强制Vue更新UI
          this.$forceUpdate()
        })
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
    async getPdfPageCount(pdfBlob) {
      this.calculatingPages = true
      
      try {
        // 使用简单的同步方法获取PDF页数
        return new Promise((resolve, reject) => {
          const fileReader = new FileReader()
          fileReader.onload = function() {
            const typedarray = new Uint8Array(this.result)
            
            // 使用pdf.js库来解析PDF并获取页数
            // 使用全局PDFJS对象，避免动态导入问题
            if (window.pdfjsLib) {
              const loadingTask = window.pdfjsLib.getDocument(typedarray)
              loadingTask.promise.then(pdf => {
                this.calculatingPages = false
                resolve(pdf.numPages)
              }).catch(error => {
                console.error('Error getting PDF page count:', error)
                this.calculatingPages = false
                resolve(0)
              })
            } else {
              // 如果pdf.js未加载，使用备用方法
              console.warn('PDF.js not loaded, using fallback')
              this.calculatingPages = false
              resolve(0)
            }
          }
          fileReader.readAsArrayBuffer(pdfBlob)
        })
      } catch (error) {
        console.error('Error in getPdfPageCount:', error)
        this.calculatingPages = false
        return 0
      }
    },
    async uploadAndPrint() {
      if (!this.printer) { this.msg = '请选择打印机'; return }
      if (!this.sides) { this.msg = '请选择打印选项'; return }
      if (!this.copies || this.copies < 1) { this.msg = '请输入有效的打印份数'; return }
      if (!this.pageRange || (this.pageRange !== 'all' && this.pageRange !== 'custom')) { this.msg = '请选择打印页面范围'; return }
      if (this.pageRange === 'custom' && !this.customPageRange) { this.msg = '请输入自定义页面范围'; return }
      if (!this.selectedFile && !this.pdfBlob) { this.msg = '请选择要打印的文件'; return }
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
      form.append('sides', this.sides)
      form.append('duplex', this.sides.startsWith('two-sided') ? 'true' : 'false')
      form.append('color', this.isColor ? 'true' : 'false')
      form.append('copies', this.copies.toString())
      
      // Add page range
      if (this.pageRange === 'all') {
        form.append('pageRange', 'all')
      } else if (this.pageRange === 'custom' && this.customPageRange) {
        form.append('pageRange', this.customPageRange)
      }

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
        localStorage.setItem('last_printer', this.printer)
      } catch (e) {
        this.msg = e.message
      }
    }
  }
}
</script>

<style scoped>
.input-error, .select-error {
  border-color: #ef4444 !important;
}

.input-error:focus, .select-error:focus {
  border-color: #ef4444 !important;
  box-shadow: 0 0 0 1px #ef4444 !important;
}
</style>
