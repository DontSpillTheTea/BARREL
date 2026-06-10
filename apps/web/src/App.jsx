import { useState, useEffect } from 'react'
import sampleData from './sampleData.json'
import './App.css'

function App() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

  const buildSha = import.meta.env.VITE_BUILD_SHA || 'dev'

  const [reviewToken, setReviewToken] = useState(localStorage.getItem('BARREL_REVIEW_TOKEN') || '')
  
  useEffect(() => {
    localStorage.setItem('BARREL_REVIEW_TOKEN', reviewToken)
  }, [reviewToken])

  const getHeaders = (additionalHeaders = {}) => {
    const headers = { ...additionalHeaders }
    if (reviewToken) {
      headers['X-BARREL-REVIEW-TOKEN'] = reviewToken
    }
    return headers
  }

  const [apiHealth, setApiHealth] = useState({ state: 'checking', error: null })
  
  const [history, setHistory] = useState([])
  const [metrics, setMetrics] = useState({ total: 0, fields: {} })
  const [batchJobs, setBatchJobs] = useState([])

  const checkHealth = () => {
    setApiHealth({ state: 'checking', error: null })
    fetch(`${apiBaseUrl}/health`, { headers: getHeaders() })
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(data => setApiHealth({ state: 'online', error: null }))
      .catch(err => setApiHealth({ state: 'offline', error: err.message }))
  }

  const fetchHistory = async () => {
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews`, { headers: getHeaders() })
      if (res.ok) {
        const data = await res.json()
        const items = Array.isArray(data) ? data : (data.reviews || [])
        setHistory(items)

        // Calculate metrics
        let total = items.length
        let fieldStats = {}
        items.forEach(item => {
          const resObj = item.result || item
          if (resObj.FieldMatches) {
            Object.keys(resObj.FieldMatches).forEach(field => {
              if (!fieldStats[field]) fieldStats[field] = { attempts: 0, matches: 0 }
              fieldStats[field].attempts++
              if (resObj.FieldMatches[field]) fieldStats[field].matches++
            })
          }
        })
        setMetrics({ total, fields: fieldStats })
      }
    } catch (e) {
      console.error('Failed to fetch history', e)
    }
  }

  const [isCheckingAuth, setIsCheckingAuth] = useState(true)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loginUsername, setLoginUsername] = useState('evaluator')
  const [loginPassword, setLoginPassword] = useState('fallback-demo-password-123')
  const [loginError, setLoginError] = useState('')

  useEffect(() => {
    let isMounted = true
    const checkAuth = async () => {
      if (!reviewToken) {
        if (isMounted) {
          setIsAuthenticated(false)
          setIsCheckingAuth(false)
        }
        return
      }
      try {
        const res = await fetch(`${apiBaseUrl}/api/v1/auth/me`, { headers: getHeaders() })
        if (isMounted) {
          setIsAuthenticated(res.ok)
        }
      } catch (err) {
        if (isMounted) setIsAuthenticated(false)
      } finally {
        if (isMounted) setIsCheckingAuth(false)
      }
    }
    checkAuth()
    return () => { isMounted = false }
  }, [apiBaseUrl, reviewToken])

  useEffect(() => {
    if (isAuthenticated) {
      checkHealth()
      fetchHistory()
    }
  }, [isAuthenticated, apiBaseUrl, reviewToken])

  const handleLogin = async (e) => {
    e.preventDefault()
    setLoginError('')
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: loginUsername, password: loginPassword })
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.error || 'Login failed')
      }
      const data = await res.json()
      setReviewToken(data.token)
      setIsAuthenticated(true)
    } catch (err) {
      setLoginError(err.message)
    }
  }

  const handleLogout = async () => {
    await fetch(`${apiBaseUrl}/api/v1/auth/logout`, { method: 'POST', headers: getHeaders() }).catch(() => {})
    setReviewToken('')
    setIsAuthenticated(false)
  }

  // Form State
  const [file, setFile] = useState(null)
  const [imagePreview, setImagePreview] = useState(null)
  const [beverageType, setBeverageType] = useState('distilled_spirits')
  const [brandName, setBrandName] = useState("Stone's Throw Spirits")
  const [classType, setClassType] = useState("Rye Whiskey")
  const [alcoholContent, setAlcoholContent] = useState("46% Alc./Vol. (92 Proof)")
  const [netContents, setNetContents] = useState("750 mL")
  const [govWarning, setGovWarning] = useState(true)
  const [ocrProvider, setOcrProvider] = useState('azure_vision')

  const [loading, setLoading] = useState(false)
  const [secondReadLoading, setSecondReadLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)
  const [jobId, setJobId] = useState(null)
  const [decisionNotes, setDecisionNotes] = useState('')

  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0]
    setFile(selectedFile)
    if (selectedFile) {
      setImagePreview(URL.createObjectURL(selectedFile))
      setResult(null)
      setError(null)
      setJobId(null)
      
      const baseName = selectedFile.name.replace(/\.[^/.]+$/, "")
      if (sampleData[baseName]) {
        const d = sampleData[baseName]
        setBrandName(d.brand_name || '')
        setClassType(d.class_type || '')
        setAlcoholContent(d.alcohol_content || '')
        setNetContents(d.net_contents || '')
        setGovWarning(d.government_warning_present || false)
        setBeverageType(d.beverage_type || 'distilled_spirits')
      }
    } else {
      setImagePreview(null)
    }
  }

  const pollJobStatus = async (pollUrl, controller) => {
    const startTime = Date.now()
    const maxWaitMs = 90000

    while (Date.now() - startTime < maxWaitMs) {
      if (controller.signal.aborted) throw new Error('AbortError')

      const res = await fetch(`${apiBaseUrl}${pollUrl}`, { 
        headers: getHeaders() 
      })
      
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.error || `HTTP error ${res.status}`)
      }
      
      const data = await res.json()
      if (data.status === 'succeeded' || data.status === 'completed') {
        return data.result
      } else if (data.status === 'failed') {
        if (data.result) return data.result
        throw new Error(data.error || 'Job failed on server')
      } else if (data.status === 'timeout') {
        throw new Error('Analysis timed out on server')
      } else if (data.status === 'unknown') {
        throw new Error('Analysis status unknown')
      }
      
      // status is queued or processing
      await new Promise(r => setTimeout(r, 2000))
    }
    throw new Error('Analysis timed out on frontend (90s limit)')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!file) {
      setError("Please select a label image first.")
      return
    }

    setLoading(true)
    setError(null)
    setResult(null)
    setJobId(null)
    setDecisionNotes('')

    const formData = new FormData()
    formData.append('file', file)
    formData.append('beverage_type', beverageType)
    
    const expectedJson = {
      brand_name: brandName,
      class_type: classType,
      alcohol_content: alcoholContent,
      net_contents: netContents,
      government_warning_present: govWarning
    }
    formData.append('expected_json', JSON.stringify(expectedJson))
    formData.append('ocr_provider', ocrProvider)

    const controller = new AbortController()
    
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/labels/analyze-async`, {
        method: 'POST',
        headers: getHeaders(),
        body: formData,
        signal: controller.signal
      })

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.error || `HTTP ${res.status}`)
      }

      const initData = await res.json()
      
      if (initData.batch) {
        alert(`Batch uploaded! Submitted ${initData.jobs?.length || 'multiple'} jobs.`)
        setBatchJobs(initData.jobs || [])
        fetchHistory()
        setFile(null)
        setImagePreview(null)
      } else {
        setJobId(initData.job_id)
        await pollJobStatus(initData.poll_url, controller)
        await loadHistoricalJob({ job_id: initData.job_id })
        fetchHistory()
      }
    } catch (err) {
      if (err.message !== 'AbortError') {
        setError(err.message || 'An unknown error occurred')
      }
    } finally {
      setLoading(false)
    }
  }

  const triggerSecondRead = async () => {
    if (!jobId && !file) return
    setSecondReadLoading(true)
    setError(null)

    try {
      let res;
      if (jobId) {
        res = await fetch(`${apiBaseUrl}/api/v1/jobs/${jobId}/second-read`, {
          method: 'POST',
          headers: getHeaders()
        })
      } else {
        const formData = new FormData()
        formData.append('file', file)
        formData.append('beverage_type', beverageType)
        const expectedJson = { brand_name: brandName, class_type: classType, alcohol_content: alcoholContent, net_contents: netContents, government_warning_present: govWarning }
        formData.append('expected_json', JSON.stringify(expectedJson))
        
        res = await fetch(`${apiBaseUrl}/api/v1/labels/second-read`, {
          method: 'POST',
          headers: getHeaders(),
          body: formData
        })
      }

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.error || `HTTP ${res.status}`)
      }
      await loadHistoricalJob({ job_id: jobId })
      fetchHistory()
    } catch (err) {
      setError(err.message || "Failed to trigger AI second read")
    } finally {
      setSecondReadLoading(false)
    }
  }

  const submitDecision = async (decision) => {
    if (!jobId) return
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews/${jobId}/decision`, {
        method: 'POST',
        headers: getHeaders({ 'Content-Type': 'application/json' }),
        body: JSON.stringify({ decision, notes: decisionNotes })
      })
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}))
        throw new Error(errData.error || `HTTP ${res.status}`)
      }
      alert(`Decision '${decision}' submitted successfully!`)
      fetchHistory()
    } catch (err) {
      alert(`Failed to submit decision: ${err.message}`)
    }
  }

  const loadHistoricalJob = async (job) => {
    const id = job.job_id || job.id
    setJobId(id)
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews/${id}`, { headers: getHeaders() })
      if (!res.ok) throw new Error('Failed to load detail')
      const detail = await res.json()
      setResult(detail)
      setDecisionNotes('')
      if (detail.original_image_url) {
        setImagePreview(`${apiBaseUrl}${detail.original_image_url}?token=${reviewToken}`)
      } else {
        setImagePreview(null)
      }
    } catch(err) {
      alert(err.message)
    }
  }

  if (isCheckingAuth) {
    return <div className="login-wrapper"><div className="loading-spinner"></div></div>
  }

  if (!isAuthenticated) {
    return (
      <div className="login-wrapper">
        <div className="login-card">
          <div className="login-header">
            <h1>BARREL</h1>
            <p>TTB Evaluator Portal</p>
          </div>
          <form onSubmit={handleLogin}>
            <div className="form-group">
              <label className="form-label">Username</label>
              <input className="form-control" type="text" value={loginUsername} onChange={e => setLoginUsername(e.target.value)} required />
            </div>
            <div className="form-group">
              <label className="form-label">Password</label>
              <input className="form-control" type="password" value={loginPassword} onChange={e => setLoginPassword(e.target.value)} required />
            </div>
            <button className="btn btn-primary" style={{width: '100%'}} type="submit">Secure Login</button>
            {loginError && <div className="alert alert-error" style={{marginTop: '1rem'}}>{loginError}</div>}
          </form>
        </div>
      </div>
    )
  }

  return (
    <div className="app-shell">
      <div className="app-header">
        <div className="logo">
          <h1>BARREL</h1>
          <p>Beverage Alcohol Review & Regulatory Evidence Logger</p>
        </div>
        <div className="auth-badge">
          <span>Evaluator Mode</span>
          <button onClick={handleLogout} className="btn btn-outline" style={{padding: '0.25rem 0.75rem'}}>Logout</button>
        </div>
      </div>

      <div className="main-grid">
        <div className="card">
          <div className="card-title">New Analysis</div>
          <form onSubmit={handleSubmit} className="analysis-form">
            <div className="file-input-wrapper">
              <div className="file-drop-area">
                {file ? <strong>{file.name}</strong> : <span>Drag & Drop or Click to Upload Label/Zip</span>}
              </div>
              <input type="file" accept="image/jpeg,image/png,image/webp,application/zip,.zip" onChange={handleFileChange} />
            </div>
            {imagePreview && !result && (
              <div style={{marginTop: '1rem', display: 'flex', justifyContent: 'center'}}>
                <img src={imagePreview} alt="Selected Label" style={{maxHeight: '300px', maxWidth: '100%', borderRadius: '8px', border: '1px solid var(--border)'}} />
              </div>
            )}

            <div className="form-group">
              <label className="form-label">OCR Engine</label>
              <select className="form-control" value={ocrProvider} onChange={e => setOcrProvider(e.target.value)}>
                <option value="azure_vision">Azure Vision (Default)</option>
                <option value="ai_based">AI Based OCR (Azure OpenAI)</option>
              </select>
            </div>

            <div className="grid-2" style={{gap: '1rem', marginTop: '1rem'}}>
              <div className="form-group">
                <label className="form-label">Beverage Type</label>
                <select className="form-control" value={beverageType} onChange={e => setBeverageType(e.target.value)}>
                  <option value="distilled_spirits">Distilled Spirits</option>
                  <option value="wine">Wine</option>
                  <option value="malt_beverages">Malt Beverages</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">Brand Name</label>
                <input className="form-control" type="text" value={brandName} onChange={e => setBrandName(e.target.value)} />
              </div>
              <div className="form-group">
                <label className="form-label">Class/Type</label>
                <input className="form-control" type="text" value={classType} onChange={e => setClassType(e.target.value)} />
              </div>
              <div className="form-group">
                <label className="form-label">Alcohol Content</label>
                <input className="form-control" type="text" value={alcoholContent} onChange={e => setAlcoholContent(e.target.value)} />
              </div>
              <div className="form-group">
                <label className="form-label">Net Contents</label>
                <input className="form-control" type="text" value={netContents} onChange={e => setNetContents(e.target.value)} />
              </div>
            </div>

            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? <span className="loading-spinner" style={{width: '1rem', height: '1rem', borderWidth: '2px'}}></span> : 'Analyze Label'}
            </button>
          </form>
          {error && <div className="alert alert-error">{error}</div>}
        </div>

        {result ? (
          <div className="review-layout">
            <div className="review-left">
              <h3 className="card-title">Evidence</h3>
              <div style={{marginBottom: '1rem', fontSize: '0.85rem'}}>
                <div><strong>Filename:</strong> {result.summary?.filename || result.result?.filename}</div>
                <div><strong>Job ID:</strong> {result.summary?.job_id}</div>
                {result.summary?.batch_id && <div><strong>Batch ID:</strong> {result.summary?.batch_id}</div>}
                <div><strong>Provider:</strong> {result.summary?.ocr_provider}</div>
                <div><strong>Submitted:</strong> {result.summary?.submitted_at ? new Date(result.summary.submitted_at).toLocaleString() : '-'}</div>
              </div>

              {imagePreview ? (
                <div className="img-preview-container">
                  <img src={imagePreview} alt="Label Preview" />
                </div>
              ) : (
                <div className="no-image">No image available</div>
              )}
              
              <div style={{marginTop: '1.5rem'}}>
                <h4>Raw OCR Text</h4>
                <div style={{maxHeight: '150px', overflowY: 'auto', background: 'rgba(255,255,255,0.05)', padding: '0.5rem', borderRadius: '4px', fontSize: '0.8rem', whiteSpace: 'pre-wrap'}}>
                  {result.raw_ocr_text || result.result?.ocr_text || 'No raw text available.'}
                </div>
              </div>

              <div style={{marginTop: '1.5rem'}}>
                <h4>Decision Panel</h4>
                <textarea 
                  className="form-control" 
                  placeholder="Review notes..." 
                  value={decisionNotes} 
                  onChange={e => setDecisionNotes(e.target.value)}
                  style={{minHeight: '80px', marginBottom: '1rem'}}
                />
                <div className="grid-3" style={{gap: '0.5rem'}}>
                  <button className="btn btn-approve" onClick={() => submitDecision('approved')}>Approve</button>
                  <button className="btn btn-reject" onClick={() => submitDecision('rejected')}>Reject</button>
                  <button className="btn btn-more-info" onClick={() => submitDecision('needs_more_info')}>RFI</button>
                </div>
              </div>
            </div>
            
            <div className="review-right">
              <div className="card">
                <div className="card-title">
                  Deterministic Extraction
                  <span className={`badge ${result?.result?.overall_status === 'Pass' ? 'badge-success' : 'badge-warning'} status-badge`} style={{ marginLeft: '1rem' }}>
                    {result?.result?.overall_status}
                  </span>
                  <span className={`badge ${result?.result?.overall_status === 'Pass' ? 'badge-success' : 'badge-warning'}`}>
                    {result.result?.overall_confidence || 0}% Confidence
                  </span>
                </div>
                <div style={{marginBottom: '1rem'}}>
                  <strong style={{fontSize: '0.85rem'}}>Expected vs Extracted:</strong>
                  <div style={{display: 'flex', gap: '1rem', fontSize: '0.8rem', marginTop: '0.5rem'}}>
                    <pre style={{flex: 1, background: 'rgba(0,0,0,0.2)', padding: '0.5rem'}}>{JSON.stringify(result.result?.expected_fields, null, 2)}</pre>
                    <pre style={{flex: 1, background: 'rgba(0,0,0,0.2)', padding: '0.5rem'}}>{JSON.stringify(result.result?.extracted_fields, null, 2)}</pre>
                  </div>
                </div>

                <ul className="field-list">
                  {(result.result?.fields || []).map((field, i) => (
                    <li className="field-item" key={i}>
                      <span className="field-name">{field.field.replace(/([A-Z])/g, ' $1').trim()}</span>
                      <span className={`badge ${field.status === 'Pass' ? 'badge-success' : 'badge-error'}`}>
                        {field.status}
                      </span>
                    </li>
                  ))}
                </ul>

                {result.result?.warnings && result.result.warnings.length > 0 && (
                  <div style={{marginTop: '1.5rem'}}>
                    <h4>Regulatory Warnings</h4>
                    {result.result.warnings.map((w, i) => (
                      <div className="alert alert-error" key={i}>
                        {w}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {result.result?.ai_escalation && (
                <div className="card ai-card">
                  <div className="card-title">
                    AI Second Read
                    <span className="badge badge-info">Azure OpenAI</span>
                  </div>
                  {result.result.ai_escalation.used ? (
                    <div>
                      <p>The AI performed a deep contextual read of this label.</p>
                    {result.result.ai_escalation.findings && result.result.ai_escalation.findings.length > 0 && (
                      <ul>
                        {result.result.ai_escalation.findings.map((f, i) => <li key={i}>{f.message || f}</li>)}
                      </ul>
                    )}
                    <div>
                      <strong>Candidate Evidence:</strong>
                      <pre style={{background: 'rgba(255,255,255,0.5)', padding: '1rem', borderRadius: '4px', marginTop: '0.5rem'}}>
                        {JSON.stringify(result.result.ai_escalation.candidates, null, 2)}
                      </pre>
                    </div>
                    </div>
                  ) : (
                    <div>
                    <p>AI was not invoked. Reason: {result.result.ai_escalation.reason || result.result.ai_escalation.error}</p>
                    {result.result.ai_escalation.eligible && (
                      <button className="btn btn-primary" onClick={triggerSecondRead} disabled={secondReadLoading}>
                        {secondReadLoading ? 'Running AI...' : 'Run AI Second Read Now'}
                      </button>
                    )}
                  </div>
                  )}
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="card" style={{display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-light)', minHeight: '300px'}}>
            Select a label to review its details
          </div>
        )}

        <div className="card review-history-section">
          <div className="card-title">Review History</div>
          {batchJobs.length > 0 && (
            <div style={{marginBottom: '2rem'}}>
              <h4 style={{marginTop: 0, marginBottom: '0.5rem', color: 'var(--accent)'}}>Batch Queue</h4>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Filename</th>
                    <th>Job ID</th>
                  </tr>
                </thead>
                <tbody>
                  {batchJobs.map((b, i) => (
                    <tr key={i}>
                      <td>{b.filename}</td>
                      <td><span style={{fontFamily: 'monospace', fontSize: '0.8rem'}}>{b.job_id}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          
          <div className="history-table-shell">
            <table className="data-table history-table">
              <thead>
                <tr>
                  <th>Submitted</th>
                  <th>Filename</th>
                  <th>Provider</th>
                  <th>Job status</th>
                  <th>Review decision</th>
                  <th>Overall status</th>
                  <th>Confidence</th>
                  <th>Fields passed</th>
                  <th>Batch ID</th>
                  <th>Job ID</th>
                </tr>
              </thead>
              <tbody>
                {history.length === 0 && (
                  <tr><td colSpan="10" style={{textAlign: 'center', padding: '1rem'}}>No recent reviews found.</td></tr>
                )}
                {history.map((item, idx) => {
                  const isPass = item.overall_status === 'Pass'
                  const dateStr = item.submitted_at
                  return (
                    <tr key={idx} className="clickable-row" onClick={() => loadHistoricalJob(item)}>
                      <td>{dateStr ? new Date(dateStr).toLocaleString() : '-'}</td>
                      <td style={{fontWeight: '500'}}>{item.filename}</td>
                      <td>{item.ocr_provider}</td>
                      <td>{item.overall_status === 'Needs Review' ? 'complete' : 'complete'}</td>
                      <td>
                        {item.reviewer_decision ? (
                          <span className="badge badge-info">{item.reviewer_decision}</span>
                        ) : (
                          <span style={{color: 'var(--text-light)'}}>unreviewed</span>
                        )}
                      </td>
                      <td>
                        <span className={`badge ${isPass ? 'badge-success' : 'badge-warning'}`}>
                          {item.overall_status || 'Unknown'}
                        </span>
                      </td>
                      <td>{item.overall_confidence || 0}%</td>
                      <td>{item.field_pass_count} / {item.field_total_count}</td>
                      <td><span style={{fontFamily: 'monospace', fontSize: '0.75rem', color: 'var(--text-light)'}}>{item.batch_id || '-'}</span></td>
                      <td><span style={{fontFamily: 'monospace', fontSize: '0.75rem', color: 'var(--text-light)'}}>{item.job_id}</span></td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>

      </div>

      <footer style={{ marginTop: '3rem', textAlign: 'center', fontSize: '0.8rem', color: 'var(--text-light)', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
        Build: {buildSha}
      </footer>
    </div>
  )
}

export default App
