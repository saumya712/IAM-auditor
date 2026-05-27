import { useState, FormEvent } from 'react'
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter'
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json'
import { githubGist } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import api from '../api/client'
import LoadingSpinner from '../components/LoadingSpinner'

SyntaxHighlighter.registerLanguage('json', json)

interface AdvisorResult {
  policy: object
  explanation: string
}

export default function RoleAdvisorView() {
  const [prompt, setPrompt] = useState('')
  const [result, setResult] = useState<AdvisorResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [validationMsg, setValidationMsg] = useState('')

  const handleGenerate = async (e: FormEvent) => {
    e.preventDefault()
    setValidationMsg('')
    setError('')

    if (!prompt.trim()) {
      setValidationMsg('Please enter a prompt before generating.')
      return
    }

    setLoading(true)
    setResult(null)
    try {
      const res = await api.post('/api/iam/advise', { prompt })
      setResult(res.data)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <h1 className="text-xl font-bold text-gray-900 mb-1">Role Advisor</h1>
      <p className="text-sm text-gray-500 mb-6">
        Describe a job role or AWS service and get a least-privilege IAM policy.
      </p>

      <form onSubmit={handleGenerate} className="mb-6">
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder="e.g. A Lambda function that reads from S3 and writes to DynamoDB"
          rows={4}
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
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
          {loading ? 'Generating…' : 'Generate'}
        </button>
      </form>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
          {error}
        </div>
      )}

      {result && (
        <div className="grid grid-cols-2 gap-4">
          {/* Policy JSON */}
          <div className="border border-gray-200 rounded-xl overflow-hidden">
            <div className="bg-gray-50 px-4 py-2 border-b border-gray-200">
              <span className="text-xs font-semibold text-gray-600 uppercase tracking-wide">
                Generated Policy
              </span>
            </div>
            <SyntaxHighlighter
              language="json"
              style={githubGist}
              customStyle={{ margin: 0, padding: '1rem', fontSize: '0.8rem', maxHeight: '480px', overflowY: 'auto' }}
            >
              {JSON.stringify(result.policy, null, 2)}
            </SyntaxHighlighter>
          </div>

          {/* Explanation */}
          <div className="border border-gray-200 rounded-xl overflow-hidden">
            <div className="bg-gray-50 px-4 py-2 border-b border-gray-200">
              <span className="text-xs font-semibold text-gray-600 uppercase tracking-wide">
                Explanation
              </span>
            </div>
            <div className="p-4 text-sm text-gray-700 leading-relaxed whitespace-pre-wrap max-h-[480px] overflow-y-auto">
              {result.explanation}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
