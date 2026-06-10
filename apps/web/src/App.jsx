import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

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
  const [ocrHealth, setOcrHealth] = useState({ state: 'checking', error: null })
  const [ocrReady, setOcrReady] = useState({ status: 'checking', details: null, error: null })

  const checkHealth = () => {
    setApiHealth({ state: 'checking', error: null })
    setOcrHealth({ state: 'checking', error: null })
    setOcrReady({ status: 'checking', details: null, error: null })

    fetch(`${apiBaseUrl}/health`, { headers: getHeaders() })
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(data => setApiHealth({ state: 'online', error: null }))
      .catch(err => setApiHealth({ state: 'offline', error: err.message }))

    fetch(`${apiBaseUrl}/health/ocr-worker`, { headers: getHeaders() })
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(data => setOcrHealth({ state: 'online', error: null }))
      .catch(err => setOcrHealth({ state: 'offline', error: err.message }))

    fetch(`${apiBaseUrl}/health/ocr-worker-ready`, { headers: getHeaders() })
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(data => setOcrReady({ status: data.status, details: data, error: null }))
      .catch(err => setOcrReady({ status: 'offline', details: null, error: err.message }))
  }

  const [history, setHistory] = useState([])
  const fetchHistory = async () => {
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/reviews`, { headers: getHeaders() })
      if (res.ok) {
        const data = await res.json()
        setHistory(Array.isArray(data) ? data : (data.reviews || []))
      }
    } catch (e) {
      console.error('Failed to fetch history', e)
    }
  }

  useEffect(() => {
    checkHealth()
    fetchHistory()
  }, [apiBaseUrl, reviewToken])

  // Form State
  const [file, setFile] = useState(null)
  const [imagePreview, setImagePreview] = useState(null)
  const [beverageType, setBeverageType] = useState('distilled_spirits')
  const [brandName, setBrandName] = useState("Stone's Throw Spirits")
  const [classType, setClassType] = useState("Rye Whiskey")
  const [alcoholContent, setAlcoholContent] = useState("46% Alc./Vol. (92 Proof)")
  const [netContents, setNetContents] = useState("750 mL")
  const [govWarning, setGovWarning] = useState(true)
  const [ocrProvider, setOcrProvider] = useState('paddleocr')

  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const paddleState = ocrReady.details?.providers?.find(p => p.provider === "paddleocr")?.state
  const isPaddleWarming = paddleState === "initializing" || ocrReady.status === "warming"
  const isPaddleReady = paddleState === "ready"
  const paddleError = ocrReady.details?.providers?.find(p => p.provider === "paddleocr")?.last_error
  const paddleDuration = ocrReady.details?.providers?.find(p => p.provider === "paddleocr")?.warmup_ms
  
  let analyzeButtonLabel = loading ? 'Analyzing...' : 'Analyze'
  let analyzeDisabled = loading
  
  if (ocrProvider === "paddleocr") {
    if (isPaddleWarming) {
      analyzeButtonLabel = 'Waiting for accurate OCR...'
      analyzeDisabled = true
    } else if (!isPaddleReady) {
      analyzeButtonLabel = 'Accurate OCR unavailable'
      analyzeDisabled = true
    }
  }

  const [jobId, setJobId] = useState(null)
  const [elapsedTime, setElapsedTime] = useState(0)
  const [decisionNotes, setDecisionNotes] = useState('')

  useEffect(() => {
    let timer
    if (loading) {
      timer = setInterval(() => {
        setElapsedTime(prev => prev + 1)
      }, 1000)
    } else {
      setElapsedTime(0)
    }
    return () => clearInterval(timer)
  }, [loading])

  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0]
    setFile(selectedFile)
    if (selectedFile) {
      setImagePreview(URL.createObjectURL(selectedFile))
    } else {
      setImagePreview(null)
    }
  }

  const pollJobStatus = async (pollUrl, controller) => {
    const startTime = Date.now()
    const maxWaitMs = 120000

    while (Date.now() - startTime < maxWaitMs) {
      if (controller.signal.aborted) throw new Error('AbortError')

      const res = await fetch(`${apiBaseUrl}${pollUrl}`, { 
        headers: getHeaders(),
        signal: controller.signal 
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      
      const job = await res.json()
      
      if (job.status === 'succeeded') {
        return job.result
      } else if (job.status === 'failed') {
        throw new Error(job.error || 'Job failed')
      }
      
      await new Promise(resolve => setTimeout(resolve, 2000))
    }
    throw new Error('Timeout: Job took longer than 120 seconds')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!file) {
      setError("Please select an image file.")
      return
    }

    if (ocrProvider === "paddleocr" && !isPaddleReady) {
      setError(`Cannot select PaddleOCR. It is currently: ${paddleState || 'unavailable'}`)
      return
    }

    setLoading(true)
    setError(null)
    setResult(null)
    setJobId(null)
    setDecisionNotes('')

    const expectedJson = {
      brand_name: brandName,
      class_type: classType,
      alcohol_content: alcoholContent,
      net_contents: netContents,
      government_warning_present: govWarning
    }

    const formData = new FormData()
    formData.append('file', file)
    formData.append('beverage_type', beverageType)
    formData.append('expected_json', JSON.stringify(expectedJson))
    formData.append('ocr_provider', ocrProvider)

    const controller = new AbortController()
    
    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/labels/analyze-async`, {
        method: 'POST',
        headers: getHeaders(), // Don't set Content-Type for FormData
        body: formData,
        signal: controller.signal
      })

      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.error || `HTTP ${res.status}`)
      }
      
      setJobId(data.job_id)
      
      const resultData = await pollJobStatus(data.poll_url, controller)
      setResult(resultData)
      fetchHistory()
    } catch (err) {
      if (err.name === 'AbortError' || err.message === 'AbortError') {
        setError("Request was cancelled.")
      } else {
        setError(err.message)
      }
    } finally {
      setLoading(false)
    }
  }

  const handleDecision = async (decision) => {
    if (!jobId) return;
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

  const loadHistoricalJob = (job) => {
    setImagePreview(null) // We may not have the image for historical jobs
    setJobId(job.job_id || job.id)
    setResult(job.result || job)
    setDecisionNotes('')
  }

  return (
    <div className="container">
      <header>
        <h1>BARREL</h1>
        <h2>Beverage Alcohol Review & Regulatory Evidence Logger</h2>
      </header>

      <div className="disclaimer">
        BARREL is a review assistant, not a final legal determination system.
        <br/>
        BARREL stores review evidence for this deployment. Uploaded labels may contain business-sensitive information.
        <br/>
        BARREL prioritizes accurate local OCR. Fast fallback is available for diagnostics, but it is not the default evidence path.
      </div>

      <div className="token-section">
        <label>BARREL_REVIEW_TOKEN: </label>
        <input 
          type="text" 
          value={reviewToken} 
          onChange={e => setReviewToken(e.target.value)} 
          placeholder="Enter token for API access..."
        />
      </div>

      <div className="status-grid">
        <div className="status-card">
          <h3>API Status</h3>
          <p className={`status-${apiHealth.state}`}>{apiHealth.state}</p>
          {apiHealth.error && <small className="error-detail">{apiHealth.error}</small>}
        </div>
        <div className="status-card">
          <h3>OCR Process</h3>
          <p className={`status-${ocrHealth.state}`}>{ocrHealth.state}</p>
          {ocrHealth.error && <small className="error-detail">{ocrHealth.error}</small>}
        </div>
        <div className="status-card">
          <h3>Accurate OCR (PaddleOCR)</h3>
          <p className={`status-${paddleState || 'checking'}`}>
            {paddleState || 'checking'}
            {isPaddleWarming ? ' (warming)' : ''}
          </p>
          {paddleDuration > 0 && <small className="success-detail">Warmup: {paddleDuration}ms</small>}
          {paddleError && <small className="error-detail">{paddleError}</small>}
        </div>
      </div>
      
      {ocrReady.details && (
        <div className="provider-status">
          <h4>Provider States</h4>
          <ul>
            <li><strong>Requires Ready for Analysis:</strong> {ocrReady.details.requires_ready_for_analysis ? 'Yes' : 'No'}</li>
            {ocrReady.details.providers?.map((p, idx) => (
              <li key={idx}><strong>{p.provider}:</strong> {p.state} <small>({p.message})</small></li>
            ))}
          </ul>
        </div>
      )}

      <div className="health-controls">
        <p>API Base URL: {apiBaseUrl}</p>
        <button onClick={() => { checkHealth(); fetchHistory(); }}>Refresh health & history</button>
      </div>

      <hr />

      <section className="history-section">
        <h3>Review History</h3>
        {history.length === 0 ? (
          <p>No previous reviews found.</p>
        ) : (
          <ul className="history-list">
            {history.map((job, idx) => (
              <li key={idx} className="history-item" onClick={() => loadHistoricalJob(job)}>
                <span><strong>Job ID:</strong> {job.job_id || job.id}</span>
                <span><strong>File:</strong> {job.filename || (job.result && job.result.filename) || 'Unknown'}</span>
                <span>
                  <span className={`status-badge ${(job.overall_status || (job.result && job.result.overall_status) || 'unknown').replace(/\s+/g, '-').toLowerCase()}`}>
                    {job.overall_status || (job.result && job.result.overall_status) || 'Unknown'}
                  </span>
                </span>
                {job.decision && <span className="decision-badge">Decision: {job.decision}</span>}
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="analysis-section">
        <h3>Single-Image Analysis</h3>
        <form onSubmit={handleSubmit} className="analysis-form">
          <div className="form-group">
            <label>Image File:</label>
            <input type="file" accept="image/*" onChange={handleFileChange} />
          </div>

          <div className="form-group">
            <label>Beverage Type:</label>
            <select value={beverageType} onChange={e => setBeverageType(e.target.value)}>
              <option value="distilled_spirits">distilled_spirits</option>
              <option value="wine">wine</option>
              <option value="malt_beverages">malt_beverages</option>
            </select>
          </div>

          <div className="form-group">
            <label>OCR Provider:</label>
            <select value={ocrProvider} onChange={e => setOcrProvider(e.target.value)}>
              <option value="paddleocr" disabled={!isPaddleReady}>Accurate local OCR (PaddleOCR)</option>
              <option value="auto">Auto accuracy mode</option>
              <option value="tesseract">Fast fallback OCR (Tesseract)</option>
            </select>
            {ocrProvider === 'tesseract' && (
              <small className="warning-text" style={{color: 'orange'}}>
                Fast fallback OCR may miss or corrupt label text. Results require review.
              </small>
            )}
          </div>

          <div className="form-group">
            <label>Expected Brand Name:</label>
            <input type="text" value={brandName} onChange={e => setBrandName(e.target.value)} />
          </div>

          <div className="form-group">
            <label>Expected Class/Type:</label>
            <input type="text" value={classType} onChange={e => setClassType(e.target.value)} />
          </div>

          <div className="form-group">
            <label>Expected Alcohol Content:</label>
            <input type="text" value={alcoholContent} onChange={e => setAlcoholContent(e.target.value)} />
          </div>

          <div className="form-group">
            <label>Expected Net Contents:</label>
            <input type="text" value={netContents} onChange={e => setNetContents(e.target.value)} />
          </div>

          <div className="form-group checkbox-group">
            <label>
              <input type="checkbox" checked={govWarning} onChange={e => setGovWarning(e.target.checked)} />
              Government Warning Expected
            </label>
          </div>

          <button type="submit" disabled={analyzeDisabled}>{analyzeButtonLabel}</button>
          <div className="form-info">
            <small>Accurate OCR can take longer on local CPU. BARREL processes it as a background job so the browser does not time out.</small>
          </div>
        </form>

        {error && <div className="error-box">{error}</div>}
        {loading && (
          <div className="loading-box">
            <p>Processing accurate OCR...</p>
            {jobId && <small>Job ID: {jobId}</small>}
            <p>Elapsed Time: {elapsedTime} seconds</p>
          </div>
        )}

        {result && (
          <div className="review-layout">
            <div className="review-left">
              <h4>Original Image</h4>
              {imagePreview ? (
                <img src={imagePreview} alt="Original Label" className="img-preview" />
              ) : (
                <div className="no-image">No image preview available</div>
              )}

              {jobId && (
                <div className="decision-controls">
                  <h4>Reviewer Decision</h4>
                  <textarea 
                    placeholder="Enter decision notes here..." 
                    value={decisionNotes} 
                    onChange={e => setDecisionNotes(e.target.value)}
                  />
                  <div className="decision-buttons">
                    <button type="button" onClick={() => handleDecision('approved')} className="btn-approve">Approve</button>
                    <button type="button" onClick={() => handleDecision('rejected')} className="btn-reject">Reject</button>
                    <button type="button" onClick={() => handleDecision('needs_more_info')} className="btn-more-info">Needs Info</button>
                  </div>
                </div>
              )}
            </div>

            <div className="review-right">
              <h4>Overall Summary</h4>
              <ul>
                <li><strong>Filename:</strong> {result.filename}</li>
                <li><strong>Beverage Type:</strong> {result.beverage_type}</li>
                <li><strong>Status:</strong> <span className={`status-badge ${result.overall_status?.replace(/\s+/g, '-').toLowerCase()}`}>{result.overall_status}</span></li>
                <li><strong>Confidence:</strong> {result.overall_confidence}</li>
              </ul>
              {result.warnings?.length > 0 && (
                <div className="warning-box">
                  <strong>Warnings:</strong>
                  <ul>
                    {result.warnings.map((w, i) => <li key={i}>{w}</li>)}
                  </ul>
                </div>
              )}

              <h4>Field Checks</h4>
              <div className="table-responsive">
                <table className="results-table">
                  <thead>
                    <tr>
                      <th>Field</th>
                      <th>Expected</th>
                      <th>Found</th>
                      <th>Status</th>
                      <th>Confidence</th>
                      <th>Explanation</th>
                      <th>Rule Citation</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.fields?.map((f, i) => (
                      <tr key={i}>
                        <td>{f.field}</td>
                        <td>{f.expected}</td>
                        <td>{f.found}</td>
                        <td><span className={`status-badge ${f.status?.replace(/\s+/g, '-').toLowerCase()}`}>{f.status}</span></td>
                        <td>{f.confidence}</td>
                        <td>{f.explanation}</td>
                        <td>
                          {f.rule && f.rule.citation ? (
                            <a href={f.rule.source_url} target="_blank" rel="noreferrer">{f.rule.citation}</a>
                          ) : 'N/A'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <h4>OCR Evidence</h4>
              {result.ocr?.status === "error" ? (
                <div className="error-box">
                  <strong>Error: </strong> {result.ocr.message} ({result.ocr.error_code})
                </div>
              ) : (
                <ul>
                  <li><strong>OCR Engine (Requested):</strong> {result.ocr?.ocr_engine}</li>
                  <li><strong>Selected Provider:</strong> {result.ocr?.selected_provider}</li>
                  <li><strong>Selection Reason:</strong> {result.ocr?.selection_reason}</li>
                  <li><strong>Mean Confidence:</strong> {result.ocr?.mean_confidence}</li>
                  <li><strong>Provider Results:</strong> 
                    <ul>
                      {result.ocr?.provider_results?.map((pr, idx) => (
                        <li key={idx}>{pr.provider}: conf {pr.mean_confidence}, len {pr.text_length}</li>
                      ))}
                    </ul>
                  </li>
                  <li><strong>Image:</strong> {result.ocr?.image_quality?.width}x{result.ocr?.image_quality?.height}</li>
                  <li><strong>Contrast Score:</strong> {result.ocr?.image_quality?.contrast_score}</li>
                  <li><strong>Blur Score:</strong> {result.ocr?.image_quality?.blur_score}</li>
                </ul>
              )}
              {result.ocr?.text && (
                <details>
                  <summary>View Raw OCR Text</summary>
                  <pre>{result.ocr.text}</pre>
                </details>
              )}

              <h4>AI Escalation Metadata</h4>
              <p><em>BARREL runs local OCR first. AI escalation is metadata-only and disabled unless explicitly configured later.</em></p>
              <ul>
                <li><strong>Eligible:</strong> {result.ai_escalation?.eligible ? 'Yes' : 'No'}</li>
                <li><strong>Used:</strong> {result.ai_escalation?.used ? 'Yes' : 'No'}</li>
                <li><strong>Provider:</strong> {result.ai_escalation?.provider}</li>
                <li><strong>Reason:</strong> {result.ai_escalation?.reason}</li>
              </ul>
            </div>
          </div>
        )}
      </section>
    </div>
  )
}

export default App
