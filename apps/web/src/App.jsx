import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

  const [apiHealth, setApiHealth] = useState({ state: 'checking', error: null })
  const [ocrHealth, setOcrHealth] = useState({ state: 'checking', error: null })

  const checkHealth = () => {
    setApiHealth({ state: 'checking', error: null })
    setOcrHealth({ state: 'checking', error: null })

    fetch(`${apiBaseUrl}/health`)
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(data => setApiHealth({ state: 'online', error: null }))
      .catch(err => setApiHealth({ state: 'offline', error: err.message }))

    fetch(`${apiBaseUrl}/health/ocr-worker`)
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        return res.json()
      })
      .then(data => setOcrHealth({ state: 'online', error: null }))
      .catch(err => setOcrHealth({ state: 'offline', error: err.message }))
  }

  useEffect(() => {
    checkHealth()
  }, [apiBaseUrl])

  // Form State
  const [file, setFile] = useState(null)
  const [beverageType, setBeverageType] = useState('distilled_spirits')
  const [brandName, setBrandName] = useState("Stone's Throw Spirits")
  const [classType, setClassType] = useState("Rye Whiskey")
  const [alcoholContent, setAlcoholContent] = useState("46% Alc./Vol. (92 Proof)")
  const [netContents, setNetContents] = useState("750 mL")
  const [govWarning, setGovWarning] = useState(true)
  const [ocrProvider, setOcrProvider] = useState('auto')

  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!file) {
      setError("Please select an image file.")
      return
    }

    setLoading(true)
    setError(null)
    setResult(null)

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

    try {
      const res = await fetch(`${apiBaseUrl}/api/v1/labels/analyze`, {
        method: 'POST',
        body: formData
      })

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }

      const data = await res.json()
      setResult(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container">
      <header>
        <h1>BARREL</h1>
        <h2>Beverage Alcohol Review & Regulatory Evidence Logger</h2>
      </header>

      <div className="disclaimer">
        BARREL is a review assistant, not a final legal determination system.
      </div>

      <div className="status-grid">
        <div className="status-card">
          <h3>API Status</h3>
          <p className={`status-${apiHealth.state}`}>{apiHealth.state}</p>
          {apiHealth.error && <small className="error-detail">{apiHealth.error}</small>}
        </div>
        <div className="status-card">
          <h3>OCR Worker Status</h3>
          <p className={`status-${ocrHealth.state}`}>{ocrHealth.state}</p>
          {ocrHealth.error && <small className="error-detail">{ocrHealth.error}</small>}
        </div>
      </div>
      <div className="health-controls">
        <p>API Base URL: {apiBaseUrl}</p>
        <button onClick={checkHealth}>Refresh health</button>
      </div>

      <hr />

      <section className="analysis-section">
        <h3>Single-Image Analysis</h3>
        <form onSubmit={handleSubmit} className="analysis-form">
          <div className="form-group">
            <label>Image File:</label>
            <input type="file" accept="image/*" onChange={e => setFile(e.target.files[0])} />
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
              <option value="auto">Auto local OCR</option>
              <option value="tesseract">Tesseract only</option>
              <option value="paddleocr">PaddleOCR only</option>
            </select>
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

          <button type="submit" disabled={loading}>{loading ? 'Analyzing...' : 'Analyze'}</button>
        </form>

        {error && <div className="error-box">{error}</div>}

        {result && (
          <div className="results-container">
            <h4>Overall Summary</h4>
            <ul>
              <li><strong>Filename:</strong> {result.filename}</li>
              <li><strong>Beverage Type:</strong> {result.beverage_type}</li>
              <li><strong>Status:</strong> <span className={`status-badge ${result.overall_status?.replace(/\s+/g, '-').toLowerCase()}`}>{result.overall_status}</span></li>
              <li><strong>Confidence:</strong> {result.overall_confidence}</li>
            </ul>

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
            <details>
              <summary>View Raw OCR Text</summary>
              <pre>{result.ocr?.text}</pre>
            </details>

            <h4>AI Escalation Metadata</h4>
            <p><em>BARREL runs local OCR first. AI escalation is metadata-only and disabled unless explicitly configured later.</em></p>
            <ul>
              <li><strong>Eligible:</strong> {result.ai_escalation?.eligible ? 'Yes' : 'No'}</li>
              <li><strong>Used:</strong> {result.ai_escalation?.used ? 'Yes' : 'No'}</li>
              <li><strong>Provider:</strong> {result.ai_escalation?.provider}</li>
              <li><strong>Reason:</strong> {result.ai_escalation?.reason}</li>
            </ul>
          </div>
        )}
      </section>
    </div>
  )
}

export default App
