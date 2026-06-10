import fs from 'fs';

async function testApi() {
  const tokenReq = await fetch('https://barrel-api.gentlesand-de48fa41.eastus.azurecontainerapps.io/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'evaluator', password: 'fallback-demo-password-123' })
  });
  const tokenData = await tokenReq.json();
  const token = tokenData.token;

  const expectedJson = {
    brand_name: 'OLD TOM DISTILLERY',
    class_type: 'Kentucky Straight Bourbon Whiskey',
    alcohol_content: '45% Alc./Vol. (90 Proof)',
    net_contents: '750 mL',
    government_warning_present: true
  };

  const formData = new FormData();
  formData.append('beverage_type', 'distilled_spirits');
  formData.append('expected_json', JSON.stringify(expectedJson));
  formData.append('ocr_provider', 'azure_vision');

  const fileBlob = new Blob([fs.readFileSync('samples/generated/good/good_01_distilled_spirits_clean_front.png')], { type: 'image/png' });
  formData.append('file', fileBlob, 'good_01_distilled_spirits_clean_front.png');

  const res = await fetch('https://barrel-api.gentlesand-de48fa41.eastus.azurecontainerapps.io/api/v1/labels/analyze-async', {
    method: 'POST',
    headers: { 'X-BARREL-REVIEW-TOKEN': token },
    body: formData
  });

  const data = await res.json();
  console.log('Analyze Response:', data);

  if (data.job_id) {
    let status = 'processing';
    let result = null;
    while (status === 'processing' || status === 'queued') {
      await new Promise(r => setTimeout(r, 2000));
      const pollRes = await fetch(`https://barrel-api.gentlesand-de48fa41.eastus.azurecontainerapps.io${data.poll_url}`, {
        headers: { 'X-BARREL-REVIEW-TOKEN': token }
      });
      const pollData = await pollRes.json();
      console.log('Poll Status:', pollData.status);
      status = pollData.status;
      if (pollData.result) result = pollData.result;
    }
    console.log('Final Result overall_status:', result?.overall_status);
    console.log('Expected Fields:', result?.expected_fields);
    console.log('Extracted Fields:', result?.extracted_fields);
    console.log('Field Check Results:', result?.fields);
  }
}

testApi().catch(console.error);
