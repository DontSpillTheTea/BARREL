import { useState, useEffect, type FormEvent, type ChangeEvent, type ReactNode } from 'react'
import './App.css'

interface GovWarningDiff {
  is_exact_match: boolean
  extracted_text: string
  canonical_text: string
  similarity: number
}

interface RuleBreadcrumb {
  id: string
  citation: string
  source_url: string
}

interface FieldCheckResult {
  field: string
  expected: string | boolean | null
  found: string | boolean | null
  status: string
  confidence: number
  similarity: number
  ai_confidence: number
  ocr_confidence: number
  parser_confidence: number
  source: string
  explanation: string
  rule: RuleBreadcrumb
  gov_warning_diff?: GovWarningDiff
}

interface PipelineTimings {
  ocr_time_ms: number
  text_parse_time_ms: number
  validation_time_ms: number
  ai_native_time_ms: number
  total_time_ms: number
}

interface AnalysisResult {
  filename: string
  requested_provider: string
  beverage_type: string
  overall_status: string
  overall_confidence: number
  processing_time_ms: number
  provider_path: string
  escalated: boolean
  escalation_reasons: string[]
  timings?: PipelineTimings
  fields: FieldCheckResult[]
  image_quality_flags?: string[]
  warnings?: string[]
  ocr_text?: string
}

interface ReviewSummary {
  id: string
  job_id: string
  filename: string
  submitted_at: string
  completed_at?: string
  provider_requested?: string
  provider_used: string
  overall_status: string
  overall_confidence: number
  field_pass_count: number
  field_total_count: number
  reviewer_decision?: string
  beverage_type: string
  brand_name?: string
  class_type?: string
  alcohol_content?: string
  net_contents?: string
}

interface ReviewDetail {
  summary: ReviewSummary
  result: AnalysisResult | null
  original_image_url?: string
  raw_ocr_text?: string
}

interface BatchJob {
  job_id: string
  filename: string
}

interface ExpectedFields {
  brand_name: string
  class_type: string
  alcohol_content: string
  net_contents: string
  government_warning_present: boolean
  producer_bottler: string
  country_of_origin: string
  beverage_type: string
}

const STATUS_BADGE: Record<string, string> = {
  'Match': 'badge-success',
  'Mismatch': 'badge-error',
  'Missing on Label': 'badge-error',
  'Missing in Application Data': 'badge-warning',
  'Uncertain': 'badge-warning',
  'Pass': 'badge-success',
  'Needs Review': 'badge-warning',
  'Likely Fail': 'badge-error',
}

const OVERALL_MATCH = ['Match', 'Pass']

const PROVIDER_PATH_LABELS: Record<string, { label: string; cls: string }> = {
  'ocr_only': { label: 'OCR Fast Path', cls: 'badge-success' },
  'ocr_then_ai_native': { label: 'OCR + AI Vision', cls: 'badge-info' },
  'ai_native_only': { label: 'AI Native', cls: 'badge-warning' },
}

function App() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
  const buildSha = import.meta.env.VITE_BUILD_SHA || 'dev'

  const [reviewToken, setReviewToken] = useState(localStorage.getItem('BARREL_REVIEW_TOKEN') || '')
  useEffect(() => { localStorage.setItem('BARREL_REVIEW_TOKEN', reviewToken) }, [reviewToken])

  const getHeaders = (extra: Record<string, string> = {}): Record<string, string> => {
    const h = { ...extra }
    if (reviewToken) h['X-BARREL-REVIEW-TOKEN'] = reviewToken
    return h
  }

  const [history, setHistory] = useState<ReviewSummary[]>([])
  const [batchJobs, setBatchJobs] = useState<BatchJob[]>([])

  const fetchHistory = async () => {
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews`, { headers: getHeaders() })
      if (res.ok) {
        const data = await res.json()
        setHistory(Array.isArray(data) ? data : (data.reviews || []))
      }
    } catch (e) { console.error('Failed to fetch history', e) }
  }

  const [isCheckingAuth, setIsCheckingAuth] = useState(true)
  const [isConnecting, setIsConnecting] = useState(false)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loginUsername, setLoginUsername] = useState('evaluator')
  const [loginPassword, setLoginPassword] = useState('')
  const [loginError, setLoginError] = useState('')

  useEffect(() => {
    let m = true
    const connectTimer = setTimeout(() => { if (m) setIsConnecting(true) }, 2000)
    const check = async () => {
      if (!reviewToken) { if (m) { setIsAuthenticated(false); setIsCheckingAuth(false) } return }
      try {
        const res = await fetch(`${apiBaseUrl}/api/v1/auth/me`, { headers: getHeaders() })
        if (m) setIsAuthenticated(res.ok)
      } catch { if (m) setIsAuthenticated(false) }
      finally { if (m) { setIsCheckingAuth(false); setIsConnecting(false) } }
    }
    check()
    return () => { m = false; clearTimeout(connectTimer) }
  }, [apiBaseUrl, reviewToken])

  useEffect(() => { if (isAuthenticated) fetchHistory() }, [isAuthenticated, apiBaseUrl, reviewToken])

  const handleLogin = async (e: FormEvent) => {
    e.preventDefault()
    setLoginError('')
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/auth/login`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: loginUsername, password: loginPassword })
      })
      if (!res.ok) { const d = await res.json().catch(() => ({} as Record<string, string>)); throw new Error(d.error || 'Login failed') }
      const data = await res.json()
      setReviewToken(data.token)
      setIsAuthenticated(true)
    } catch (err) { setLoginError((err as Error).message) }
  }

  const handleLogout = async () => {
    await fetch(`${apiBaseUrl}/api/v1/auth/logout`, { method: 'POST', headers: getHeaders() }).catch(() => {})
    setReviewToken('')
    setIsAuthenticated(false)
  }

  const [file, setFile] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ReviewDetail | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [jobId, setJobId] = useState<string | null>(null)
  const [decisionNotes, setDecisionNotes] = useState('')
  const [showExpectedFields, setShowExpectedFields] = useState(false)
  const [expandedGovWarning, setExpandedGovWarning] = useState(false)

  const [expectedFields, setExpectedFields] = useState<ExpectedFields>({
    brand_name: '', class_type: '', alcohol_content: '', net_contents: '',
    government_warning_present: true, producer_bottler: '', country_of_origin: '',
    beverage_type: 'distilled_spirits',
  })

  const updateExpected = (key: keyof ExpectedFields, val: string | boolean) =>
    setExpectedFields(prev => ({ ...prev, [key]: val }))

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0] || null
    setFile(f)
    if (f) { setImagePreview(URL.createObjectURL(f)); setResult(null); setError(null); setJobId(null) }
    else setImagePreview(null)
  }

  const pollJobStatus = async (pollUrl: string, controller: AbortController): Promise<AnalysisResult> => {
    const start = Date.now()
    while (Date.now() - start < 90000) {
      if (controller.signal.aborted) throw new Error('AbortError')
      const res = await fetch(`${apiBaseUrl}${pollUrl}`, { headers: getHeaders() })
      if (!res.ok) {
        if (res.status === 404) { await new Promise(r => setTimeout(r, 2000)); continue }
        const d = await res.json().catch(() => ({} as Record<string, string>))
        throw new Error(d.error || `HTTP error ${res.status}`)
      }
      const data = await res.json()
      if (data.status === 'succeeded' || data.status === 'completed') return data.result
      if (data.status === 'failed') { if (data.result) return data.result; throw new Error(data.error || 'Job failed') }
      if (data.status === 'timeout' || data.status === 'unknown') throw new Error(`Analysis ${data.status}`)
      await new Promise(r => setTimeout(r, 2000))
    }
    throw new Error('Analysis timed out (90s limit)')
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!file) { setError('Please select a label image first.'); return }
    setLoading(true); setError(null); setResult(null); setJobId(null); setDecisionNotes('')

    const formData = new FormData()
    formData.append('file', file)
    formData.append('beverage_type', expectedFields.beverage_type)
    formData.append('expected_json', JSON.stringify({
      brand_name: expectedFields.brand_name,
      class_type: expectedFields.class_type,
      alcohol_content: expectedFields.alcohol_content,
      net_contents: expectedFields.net_contents,
      government_warning_present: expectedFields.government_warning_present,
      producer_bottler: expectedFields.producer_bottler,
      country_of_origin: expectedFields.country_of_origin,
    }))

    const controller = new AbortController()
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/labels/analyze-async`, {
        method: 'POST', headers: getHeaders(), body: formData, signal: controller.signal
      })
      if (!res.ok) { const d = await res.json().catch(() => ({} as Record<string, string>)); throw new Error(d.error || `HTTP ${res.status}`) }
      const initData = await res.json()

      if (initData.batch) {
        alert(`Batch uploaded! Submitted ${initData.jobs?.length || 'multiple'} jobs.`)
        setBatchJobs(initData.jobs || [])
        fetchHistory()
        setFile(null); setImagePreview(null)
      } else {
        setJobId(initData.job_id)
        await pollJobStatus(initData.poll_url, controller)
        await loadHistoricalJob({ job_id: initData.job_id } as ReviewSummary)
        fetchHistory()
      }
    } catch (err) {
      if ((err as Error).message !== 'AbortError') setError((err as Error).message || 'An unknown error occurred')
    } finally { setLoading(false) }
  }

  const submitDecision = async (decision: string) => {
    if (!jobId) return
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews/${jobId}/decision`, {
        method: 'POST', headers: getHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify({ decision, notes: decisionNotes })
      })
      if (!res.ok) { const d = await res.json().catch(() => ({} as Record<string, string>)); throw new Error(d.error || `HTTP ${res.status}`) }
      alert(`Decision '${decision}' submitted.`)
      fetchHistory()
    } catch (err) { alert(`Failed: ${(err as Error).message}`) }
  }

  const loadHistoricalJob = async (job: ReviewSummary) => {
    const id = job.job_id || job.id
    setJobId(id); setFile(null); setExpandedGovWarning(false)
    try {
      let detail: ReviewDetail | null = null
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews/${id}`, { headers: getHeaders() })
      if (res.ok) {
        detail = await res.json()
      } else {
        const jobRes = await fetch(`${apiBaseUrl}/api/v1/jobs/${id}`, { headers: getHeaders() })
        if (!jobRes.ok) throw new Error('Failed to load detail')
        const jd = await jobRes.json()
        detail = {
          summary: {
            id, job_id: id,
            filename: job.filename || jd.filename || jd.result?.filename || '',
            submitted_at: job.submitted_at || jd.created_at || '',
            completed_at: job.completed_at || jd.updated_at || '',
            provider_requested: job.provider_requested || jd.result?.requested_provider || '',
            provider_used: job.provider_used || jd.result?.ai_second_read?.provider || jd.result?.requested_provider || '',
            overall_status: job.overall_status || jd.result?.overall_status || '',
            overall_confidence: job.overall_confidence || jd.result?.overall_confidence || 0,
            field_pass_count: 0, field_total_count: 0, beverage_type: '',
          },
          result: jd.result || null,
          original_image_url: `/api/v1/reviews/${id}/image`,
          raw_ocr_text: jd.result?.ocr_text || '',
        }
      }
      setResult(detail)
      setDecisionNotes('')
      if (detail?.original_image_url) {
        setImagePreview(`${apiBaseUrl}${detail.original_image_url}?token=${reviewToken}`)
      } else { setImagePreview(null) }
    } catch (err) { setError((err as Error).message || 'Failed to load detail') }
  }

  const renderProviderBadge = (): ReactNode => {
    const path = result?.result?.provider_path
    if (!path) return null
    const info = PROVIDER_PATH_LABELS[path]
    if (info) return <span className={`badge ${info.cls}`}>{info.label}</span>
    return null
  }

  const renderTimingBreakdown = (): ReactNode => {
    const t = result?.result?.timings
    const ms = t?.total_time_ms || result?.result?.processing_time_ms
    if (!ms) return null
    const sec = (ms / 1000).toFixed(1)
    const cls = ms < 5000 ? 'badge-success' : ms < 10000 ? 'badge-warning' : 'badge-error'
    let detail = ''
    if (t) {
      const parts: string[] = []
      if (t.ocr_time_ms) parts.push(`OCR: ${(t.ocr_time_ms / 1000).toFixed(1)}s`)
      if (t.text_parse_time_ms) parts.push(`Parse: ${(t.text_parse_time_ms / 1000).toFixed(1)}s`)
      if (t.ai_native_time_ms) parts.push(`AI: ${(t.ai_native_time_ms / 1000).toFixed(1)}s`)
      if (parts.length) detail = ` (${parts.join(' | ')})`
    }
    return <span className={`badge ${cls}`} title={detail}>{sec}s{detail}</span>
  }

  const renderGovWarningDiff = (diff: GovWarningDiff): ReactNode => {
    if (diff.is_exact_match) return <div className="gov-diff-detail"><span className="badge badge-success">Exact match with statutory text</span></div>

    const chars: ReactNode[] = []
    const maxLen = Math.max(diff.canonical_text.length, diff.extracted_text.length)
    for (let i = 0; i < maxLen; i++) {
      const c = diff.canonical_text[i]
      const e = diff.extracted_text[i]
      if (c === e) chars.push(<span key={i}>{c}</span>)
      else if (e === undefined) chars.push(<span key={i} className="diff-missing">{c}</span>)
      else chars.push(<span key={i} className="diff-mismatch">{e}</span>)
    }

    return (
      <div className="gov-diff-detail">
        <div style={{ marginBottom: '0.5rem', fontSize: '0.8rem' }}><strong>Similarity:</strong> {(diff.similarity * 100).toFixed(1)}%</div>
        <div style={{ marginBottom: '0.5rem' }}>
          <strong style={{ fontSize: '0.75rem' }}>Extracted (differences highlighted):</strong>
          <div className="diff-text">{chars}</div>
        </div>
        <div>
          <strong style={{ fontSize: '0.75rem' }}>Canonical statutory text:</strong>
          <div className="diff-text" style={{ color: 'var(--text-light)' }}>{diff.canonical_text}</div>
        </div>
      </div>
    )
  }

  const formatFieldValue = (val: string | boolean | null | undefined): string => {
    if (val === undefined || val === null) return '-'
    if (typeof val === 'boolean') return val ? 'Yes' : 'No'
    return String(val)
  }

  if (isCheckingAuth) return (
    <div className="login-wrapper">
      <div style={{ textAlign: 'center' }}>
        <div className="loading-spinner"></div>
        {isConnecting && <p style={{ color: 'var(--text-light)', marginTop: '1rem', fontSize: '0.9rem' }}>Connecting to server...</p>}
      </div>
    </div>
  )

  if (!isAuthenticated) {
    return (
      <div className="login-wrapper">
        <div className="login-card">
          <div className="login-header"><h1>BARREL</h1><p>TTB Evaluator Portal</p></div>
          <form onSubmit={handleLogin}>
            <div className="form-group">
              <label className="form-label">Username</label>
              <input className="form-control" type="text" value={loginUsername} onChange={e => setLoginUsername(e.target.value)} required />
            </div>
            <div className="form-group">
              <label className="form-label">Password</label>
              <input className="form-control" type="password" value={loginPassword} onChange={e => setLoginPassword(e.target.value)} required />
            </div>
            <button className="btn btn-primary" style={{ width: '100%' }} type="submit">Secure Login</button>
            {loginError && <div className="alert alert-error" style={{ marginTop: '1rem' }}>{loginError}</div>}
          </form>
        </div>
      </div>
    )
  }

  const overallStatus = result?.result?.overall_status
  const isOverallMatch = OVERALL_MATCH.includes(overallStatus || '')

  return (
    <div className="app-shell">
      <div className="app-header">
        <div className="logo"><h1>BARREL</h1><p>Beverage Alcohol Review & Regulatory Evidence Logger</p></div>
        <div className="auth-badge">
          <span>Evaluator Mode</span>
          <button onClick={handleLogout} className="btn btn-outline" style={{ padding: '0.25rem 0.75rem' }}>Logout</button>
        </div>
      </div>

      <div className="top-row">
        <div className="card upload-panel">
          <div className="card-title">New Analysis</div>
          <form onSubmit={handleSubmit} className="analysis-form">
            <div className="file-input-wrapper">
              <div className="file-drop-area">
                {file ? <strong>{file.name}</strong> : <span>Drag & Drop or Click to Upload Label/Zip</span>}
              </div>
              <input type="file" accept="image/jpeg,image/png,image/webp,application/zip,.zip" onChange={handleFileChange} />
            </div>

            <button type="button" className="btn btn-outline toggle-fields-btn" onClick={() => setShowExpectedFields(!showExpectedFields)}>
              {showExpectedFields ? 'Hide' : 'Show'} Expected Fields {showExpectedFields ? '▲' : '▼'}
            </button>

            {showExpectedFields && (
              <div className="expected-fields-form">
                <div className="form-group"><label className="form-label">Brand Name</label><input className="form-control" type="text" value={expectedFields.brand_name} onChange={e => updateExpected('brand_name', e.target.value)} /></div>
                <div className="form-group"><label className="form-label">Class/Type</label><input className="form-control" type="text" value={expectedFields.class_type} onChange={e => updateExpected('class_type', e.target.value)} /></div>
                <div className="form-group"><label className="form-label">Alcohol Content</label><input className="form-control" type="text" value={expectedFields.alcohol_content} onChange={e => updateExpected('alcohol_content', e.target.value)} placeholder="e.g. 45% Alc./Vol." /></div>
                <div className="form-group"><label className="form-label">Net Contents</label><input className="form-control" type="text" value={expectedFields.net_contents} onChange={e => updateExpected('net_contents', e.target.value)} placeholder="e.g. 750 mL" /></div>
                <div className="form-group"><label className="form-label">Producer/Bottler <span className="optional-tag">(optional)</span></label><input className="form-control" type="text" value={expectedFields.producer_bottler} onChange={e => updateExpected('producer_bottler', e.target.value)} /></div>
                <div className="form-group"><label className="form-label">Country of Origin <span className="optional-tag">(optional)</span></label><input className="form-control" type="text" value={expectedFields.country_of_origin} onChange={e => updateExpected('country_of_origin', e.target.value)} /></div>
                <div className="form-group"><label className="form-label">Beverage Type</label>
                  <select className="form-control" value={expectedFields.beverage_type} onChange={e => updateExpected('beverage_type', e.target.value)}>
                    <option value="distilled_spirits">Distilled Spirits</option>
                    <option value="wine">Wine</option>
                    <option value="malt_beverages">Malt Beverages</option>
                  </select>
                </div>
                <div className="form-group checkbox-group"><label><input type="checkbox" checked={expectedFields.government_warning_present} onChange={e => updateExpected('government_warning_present', e.target.checked)} /><span>Government Warning Required</span></label></div>
              </div>
            )}

            <button type="submit" className="btn btn-primary" disabled={loading} style={{ width: '100%' }}>
              {loading ? <span className="loading-spinner" style={{ width: '1rem', height: '1rem', borderWidth: '2px' }}></span> : 'Analyze Upload'}
            </button>
          </form>
          {error && <div className="alert alert-error">{error}</div>}
        </div>

        <div className="card image-panel">
          {imagePreview ? (
            <div className="image-viewbox"><img src={imagePreview} alt="Label Preview" /></div>
          ) : (
            <div className="no-image">Select or upload a label image to preview</div>
          )}
          {result && (
            <div className="decision-row">
              <div className="decision-meta">
                <span className={`badge ${isOverallMatch ? 'badge-success' : 'badge-error'} status-badge`}>{overallStatus}</span>
                <span className={`badge ${isOverallMatch ? 'badge-success' : 'badge-warning'}`}>{result.result?.overall_confidence || 0}%</span>
                {renderProviderBadge()}
                {renderTimingBreakdown()}
                <span style={{ fontSize: '0.8rem', color: 'var(--text-light)' }}>{result.summary?.filename}</span>
              </div>
              {result.result?.escalated && result.result?.escalation_reasons?.length > 0 && (
                <div className="alert alert-info" style={{ margin: '0.5rem 0 0 0', padding: '0.5rem 0.75rem', fontSize: '0.8rem' }}>
                  Escalated to AI vision: {result.result.escalation_reasons.join('; ')}
                </div>
              )}
              <div className="decision-actions">
                <textarea className="form-control decision-notes" placeholder="Review notes..." value={decisionNotes} onChange={e => setDecisionNotes(e.target.value)} />
                <button className="btn btn-approve" onClick={() => submitDecision('approved')}>Approve</button>
                <button className="btn btn-reject" onClick={() => submitDecision('rejected')}>Reject</button>
              </div>
            </div>
          )}
        </div>
      </div>

      {result && (
        <div className="card field-table-full">
          <div className="card-title">
            Field Verification
            <span style={{ fontSize: '0.8rem', color: 'var(--text-light)', fontWeight: 'normal' }}>
              Provider: {result.summary?.provider_used || result.result?.requested_provider || 'unknown'}
            </span>
          </div>
          <div className="table-responsive">
            <table className="data-table" style={{ width: '100%' }}>
              <thead><tr><th>Field</th><th>Expected</th><th>Extracted from Label</th><th>Match</th><th>Status</th></tr></thead>
              <tbody>
                {(result.result?.fields || []).map((field, i) => {
                  const badgeClass = STATUS_BADGE[field.status] || 'badge-warning'
                  const isGovWarning = field.field === 'Government Warning'
                  const hasDiff = isGovWarning && field.gov_warning_diff

                  return (
                    <tr key={i} className={hasDiff ? 'expandable-row' : ''} onClick={hasDiff ? () => setExpandedGovWarning(!expandedGovWarning) : undefined}>
                      <td style={{ fontWeight: '600' }}>
                        {field.field}
                        {field.explanation && <div style={{ fontSize: '0.75rem', fontWeight: 'normal', color: 'var(--text-light)', marginTop: '0.25rem' }}>{field.explanation}</div>}
                        {field.rule?.citation && (
                          <div style={{ fontSize: '0.7rem', color: 'var(--accent)', marginTop: '0.1rem', fontWeight: 'normal' }}>
                            <a href={field.rule.source_url} target="_blank" rel="noreferrer" style={{ color: 'var(--accent)', textDecoration: 'underline' }}>{field.rule.citation}</a>
                          </div>
                        )}
                        {hasDiff && !expandedGovWarning && (
                          <div style={{ fontSize: '0.7rem', color: 'var(--warning)', marginTop: '0.2rem', cursor: 'pointer' }}>
                            {field.gov_warning_diff!.is_exact_match ? 'Exact match ✓' : `${Math.round((1 - field.gov_warning_diff!.similarity) * field.gov_warning_diff!.canonical_text.length)} character differences ▼`}
                          </div>
                        )}
                        {hasDiff && expandedGovWarning && renderGovWarningDiff(field.gov_warning_diff!)}
                      </td>
                      <td style={{ fontSize: '0.8rem' }}>{formatFieldValue(field.expected)}</td>
                      <td style={{ fontSize: '0.8rem', color: 'var(--text-bright)' }}>{formatFieldValue(field.found)}</td>
                      <td style={{ fontSize: '0.8rem' }}>
                        {field.similarity > 0 && <div style={{ fontWeight: '500' }}>{Math.round(field.similarity * 100)}% sim</div>}
                        {field.ai_confidence > 0 && <div style={{ fontSize: '0.7rem', color: field.ai_confidence < 0.7 ? 'var(--warning)' : 'var(--text-light)' }}>AI: {Math.round(field.ai_confidence * 100)}%</div>}
                        {!field.similarity && !field.ai_confidence && <span>{field.confidence || 0}%</span>}
                        {field.source && <div style={{ fontSize: '0.65rem', color: 'var(--text-light)', marginTop: '0.15rem' }}>{field.source === 'ai_native' ? 'AI Vision' : field.source === 'ocr_text' ? 'OCR' : field.source}</div>}
                      </td>
                      <td><span className={`badge ${badgeClass}`}>{field.status}</span></td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          {result.result?.image_quality_flags && result.result.image_quality_flags.length > 0 && (
            <div className="alert alert-info" style={{ marginTop: '1rem' }}>Image quality: {result.result.image_quality_flags.join(', ')}. Results may be less reliable.</div>
          )}
          {result.result?.warnings && result.result.warnings.length > 0 && (
            <div style={{ marginTop: '1rem' }}><h4>Regulatory Warnings</h4>{result.result.warnings.map((w, i) => <div className="alert alert-error" key={i}>{w}</div>)}</div>
          )}

          <div style={{ marginTop: '1rem', textAlign: 'right' }}>
            <button className="btn btn-outline" onClick={() => {
              const fields = result.result?.fields || []
              const rows: string[][] = [['Field', 'Expected', 'Extracted', 'Similarity', 'AI Confidence', 'Status', 'CFR Citation', 'Explanation']]
              fields.forEach(f => {
                rows.push([f.field, formatFieldValue(f.expected), formatFieldValue(f.found), f.similarity ? Math.round(f.similarity * 100) + '%' : '', f.ai_confidence ? Math.round(f.ai_confidence * 100) + '%' : '', f.status, f.rule?.citation || '', f.explanation || ''])
              })
              const r = result.result
              const header = `Filename,${result.summary?.filename || ''}\nOverall Status,${r?.overall_status || ''}\nProvider Path,${r?.provider_path || ''}\nEscalated,${r?.escalated || false}\nEscalation Reasons,"${(r?.escalation_reasons || []).join('; ')}"\nTotal Time,${r?.timings?.total_time_ms || r?.processing_time_ms || 0}ms\nOCR Time,${r?.timings?.ocr_time_ms || 0}ms\nAI Time,${r?.timings?.ai_native_time_ms || 0}ms\n\n`
              const csv = header + rows.map(row => row.map(c => '"' + String(c).replace(/"/g, '""') + '"').join(',')).join('\n')
              const blob = new Blob([csv], { type: 'text/csv' })
              const url = URL.createObjectURL(blob)
              const a = document.createElement('a')
              a.href = url; a.download = `barrel-verification-${result.summary?.job_id || 'report'}.csv`
              a.click(); URL.revokeObjectURL(url)
            }}>Export CSV</button>
          </div>
        </div>
      )}

      {!result && (
        <div className="card" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-light)', minHeight: '200px' }}>
          Upload a label and click Analyze to see verification results
        </div>
      )}

      <div className="card review-history-section">
        <div className="card-title">Review History</div>
        {batchJobs.length > 0 && (
          <div style={{ marginBottom: '2rem' }}>
            <h4 style={{ marginTop: 0, marginBottom: '0.5rem', color: 'var(--accent)' }}>Batch Queue</h4>
            <table className="data-table"><thead><tr><th>Filename</th><th>Job ID</th></tr></thead>
              <tbody>{batchJobs.map((b, i) => <tr key={i}><td>{b.filename}</td><td style={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>{b.job_id}</td></tr>)}</tbody>
            </table>
          </div>
        )}
        <div className="history-table-shell">
          <table className="data-table history-table">
            <thead><tr><th>Filename</th><th>Status</th><th>Confidence</th><th>Decision</th><th>Brand</th><th>Class/Type</th><th>ABV</th><th>Net Contents</th><th>Submitted</th></tr></thead>
            <tbody>
              {history.length === 0 && <tr><td colSpan={9} style={{ textAlign: 'center', padding: '1rem' }}>No recent reviews found.</td></tr>}
              {history.map((item, idx) => {
                const match = OVERALL_MATCH.includes(item.overall_status)
                return (
                  <tr key={idx} className="clickable-row" onClick={() => loadHistoricalJob(item)}>
                    <td style={{ fontWeight: '500' }}>{item.filename}</td>
                    <td><span className={`badge ${match ? 'badge-success' : 'badge-warning'}`}>{item.overall_status || 'Unknown'}</span></td>
                    <td>{item.overall_confidence || 0}%</td>
                    <td>{item.reviewer_decision ? <span className="badge badge-info">{item.reviewer_decision}</span> : <span style={{ color: 'var(--text-light)' }}>unreviewed</span>}</td>
                    <td>{item.brand_name || '-'}</td>
                    <td>{item.class_type || '-'}</td>
                    <td>{item.alcohol_content || '-'}</td>
                    <td>{item.net_contents || '-'}</td>
                    <td>{item.submitted_at ? new Date(item.submitted_at).toLocaleString() : '-'}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>

      <footer style={{ marginTop: '2rem', textAlign: 'center', fontSize: '0.8rem', color: 'var(--text-light)', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
        Build: {buildSha}
      </footer>
    </div>
  )
}

export default App
