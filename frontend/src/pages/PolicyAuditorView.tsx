import { useState, FormEvent, useRef, DragEvent } from 'react'
import api from '../api/client'
import LoadingSpinner from '../components/LoadingSpinner'
import FindingsReport from '../components/FindingsReport'

interface Finding {
  description: string
  risk_level: 'High' | 'Medium' | 'Low'
  remediation: string
}

export default function PolicyAuditorView() {
  const [policy, setPolicy] = useState('')
  const [findings, setFindings] = useState<Finding[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [validationMsg, setValidationMsg] = useState('')
  const [dragging, setDragging] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const loadFile = (file: File) => {
    const reader = new FileReader()
    reader.onload = (e) => setPolicy(e.target?.result as string)
    reader.readAsText(file)
  }

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files[0]
    if (file) loadFile(file)
  }

  const handleAudit = async (e: FormEvent) => {
    e.preventDefault()
    setValidationMsg('')
    setError('')

    if (!policy.trim()) {
      setValidationMsg('Please enter or upload an IAM policy before auditing.')
      return
    }

    setLoading(true)
    setFindings(null)
    try {
      const res = await api.post('/api/iam/audit', { policy })
      setFindings(res.data.findings)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <h1 className="text-xl font-bold text-gray-900 mb-1">Policy Auditor</h1>
      <p className="text-sm text-gray-500 mb-6">
        Paste or upload an IAM JSON policy to get a security analysis.
      </p>

      <form onSubmit={handleAudit} className="mb-6">
        {/* Dropzone */}
        <div
          onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
          onDragLeave={() => setDragging(false)}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          className={`mb-3 border-2 border-dashed rounded-xl p-4 text-center cursor-pointer transition-colors ${
            dragging ? 'border-blue-400 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
          }`}
        >
          <p className="text-sm text-gray-500">
            Drop a JSON file here or{' '}
            <span className="text-blue-600 font-medium">click to browse</span>
          </p>
          <input
            ref={fileInputRef}
            type="file"
            accept=".json,application/json"
            className="hidden"
            onChange={(e) => { const f = e.target.files?.[0]; if (f) loadFile(f) }}
          />
        </div>

        <textarea
          value={policy}
          onChange={(e) => setPolicy(e.target.value)}
          placeholder='{"Version": "2012-10-17", "Statement": [...]}'
          rows={8}
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
        />
        {validationMsg && (
          <p className="mt-1 text-xs text-red-600">{validationMsg}</p>
        )}
        <button
          type="submit"
          disabled={loading}
          className="mt-3 flex items-center gap-2 bg-blue-600 text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
        >
          {loading && <LoadingSpinner size="sm" />}
          {loading ? 'Auditing…' : 'Audit'}
        </button>
      </form>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
          {error}
        </div>
      )}

      {findings !== null && <FindingsReport findings={findings} />}
    </div>
  )
}
