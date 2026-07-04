import { readFile, writeFile } from 'node:fs/promises'

const [, , inputPath, outputPath = inputPath] = process.argv

if (!inputPath) {
  console.error('Usage: node scripts/postprocess-openapi.mjs <input.yaml> [output.yaml]')
  process.exit(1)
}

const source = await readFile(inputPath, 'utf8')
const lines = source.split('\n')
const result = []

let changed = 0

for (let index = 0; index < lines.length;) {
  const line = lines[index]
  const parameterMatch = line.match(/^(\s*)- name:/)

  if (!parameterMatch) {
    result.push(line)
    index += 1
    continue
  }

  const parameterIndent = parameterMatch[1]
  const block = [line]
  index += 1

  while (index < lines.length) {
    const nextLine = lines[index]
    const nextParameterMatch = nextLine.match(/^(\s*)- name:/)

    if (nextParameterMatch && nextParameterMatch[1].length <= parameterIndent.length) {
      break
    }

    block.push(nextLine)
    index += 1
  }

  const isQueryParameter = block.some((entry) => /^\s*in:\s*query\s*$/.test(entry))
  const isArraySchema = block.some((entry) => /^\s*type:\s*array\s*$/.test(entry))
  const hasExplode = block.some((entry) => /^\s*explode:\s*/.test(entry))

  if (isQueryParameter && isArraySchema && !hasExplode) {
    const attributeIndent = `${parameterIndent}  `

    if (!block.some((entry) => /^\s*style:\s*/.test(entry))) {
      block.push(`${attributeIndent}style: form`)
    }

    block.push(`${attributeIndent}explode: true`)
    changed += 1
  }

  result.push(...block)
}

await writeFile(outputPath, result.join('\n'))
console.log(`Added explicit repeat-style serialization to ${changed} array query parameter(s).`)
