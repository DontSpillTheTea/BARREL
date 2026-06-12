import fs from 'fs'
import path from 'path'

const walkSync = (dir: string, filelist: string[] = []): string[] => {
  if (!fs.existsSync(dir)) return filelist
  fs.readdirSync(dir).forEach(file => {
    const dirFile = path.join(dir, file)
    if (fs.statSync(dirFile).isDirectory()) {
      filelist = walkSync(dirFile, filelist)
    } else {
      filelist.push(dirFile)
    }
  })
  return filelist
}

const jsonFiles = walkSync('samples/expected').filter(f => f.endsWith('.json'))
const out: Record<string, unknown> = {}
for (const f of jsonFiles) {
  const content = fs.readFileSync(f, 'utf8')
  out[path.basename(f, '.json')] = JSON.parse(content)
}

fs.writeFileSync('apps/web/src/sampleData.json', JSON.stringify(out, null, 2))
console.log('Generated apps/web/src/sampleData.json')
