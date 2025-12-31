<template>
  <div class="p-6">
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h2 class="card-title">Print</h2>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="col-span-1 space-y-4">
            <label class="label">
              <span class="label-text">Printer</span>
            </label>
            <select v-model="printer" class="select select-bordered w-full">
              <option v-for="p in printers" :key="p.uri" :value="p.uri">{{ p.name }} — {{ p.uri }}</option>
            </select>

            <div>
              <label class="label">
                <span class="label-text">File</span>
              </label>
              <input type="file" ref="file" @change="onFileChange" class="file-input file-input-bordered w-full" />
            </div>

            <div class="space-x-2">
              <button class="btn btn-primary" :disabled="!canPrint || converting" @click="uploadAndPrint">Print</button>
              <button class="btn" :disabled="!selectedFile || converting" @click="convertToPdf">Convert</button>
              <a v-if="previewUrl" :href="previewUrl" :download="downloadName" class="btn btn-ghost">Download Preview</a>
            </div>

            <div class="text-sm text-muted">{{ msg }}</div>

            <div class="mt-4">
              <label class="label"><span class="label-text">Conversion status</span></label>
              <div v-if="converting" class="alert alert-info">Converting…</div>
              <div v-if="converted" class="alert alert-success">Converted to PDF</div>
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
              <div v-else class="p-4 text-muted">No preview available</div>
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
      downloadName: ''
    }
  },
  computed: {
    canPrint() {
      return !!this.printer && (!!this.pdfBlob || !!this.selectedFile)
    }
  },
  async mounted() {
    try {
      const resp = await fetch('/api/printers', { credentials: 'include' })
      if (resp.ok) {
        this.printers = await resp.json()
        const last = localStorage.getItem('last_printer')
        if (last) this.printer = last
        else if (this.printers.length > 0) this.printer = this.printers[0].uri
      } else {
        this.msg = 'failed to load printers'
      }
    } catch (e) {
      this.msg = 'failed to load printers: ' + e.message
    }
  },
  methods: {
    getCSRF() {
      const m = document.cookie.match('(^|;)\\s*csrf_token\\s*=\\s*([^;]+)')
      return m ? m.pop() : ''
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
    },
    async convertToPdf() {
      if (!this.selectedFile) { this.msg = 'No file selected'; return }
      this.converting = true
      this.msg = ''
      try {
        const f = this.selectedFile
        let blob = null
        if (f.type.startsWith('image/')) {
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
        this.msg = 'Conversion ready'
      } catch (e) {
        this.msg = 'Conversion failed: ' + e.message
      } finally {
        this.converting = false
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
      if (!this.printer) { this.msg = 'Select a printer'; return }
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
        this.msg = 'No file to print'
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
        if (!resp.ok) throw new Error('print failed')
        const j = await resp.json()
        this.msg = 'Job queued: ' + (j.jobId || '')
        localStorage.setItem('last_printer', this.printer)
      } catch (e) {
        this.msg = e.message
      }
    }
  }
}
</script>
