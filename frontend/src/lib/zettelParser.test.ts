import { describe, it, expect } from 'vitest'
import { stripSuggestionBlocks, parseZettelSuggestions, parseConnectionSuggestions } from './zettelParser'

describe('stripSuggestionBlocks', () => {
  it('strips ZETTEL_SUGGESTION blocks from text', () => {
    const input = 'Hello ---ZETTEL_SUGGESTION---\ntitle: Test\n---END_ZETTEL--- world'
    expect(stripSuggestionBlocks(input)).toBe('Hello  world')
  })

  it('strips CONNECTION_SUGGESTION blocks', () => {
    const input = 'Before ---CONNECTION_SUGGESTION---\nsource_title: A\ntarget_title: B\n---END_CONNECTION--- after'
    expect(stripSuggestionBlocks(input)).toBe('Before  after')
  })

  it('returns unchanged text when no blocks present', () => {
    const input = 'Just some regular text with no blocks.'
    expect(stripSuggestionBlocks(input)).toBe(input)
  })

  it('handles text before and after blocks', () => {
    const input = 'Start text\n---ZETTEL_SUGGESTION---\ntitle: X\n---END_ZETTEL---\nEnd text'
    const result = stripSuggestionBlocks(input)
    expect(result).toContain('Start text')
    expect(result).toContain('End text')
    expect(result).not.toContain('ZETTEL_SUGGESTION')
  })
})

describe('parseZettelSuggestions', () => {
  it('parses a single suggestion with all fields', () => {
    const input = `---ZETTEL_SUGGESTION---
title: Dependency Injection
type: concept
summary: A design pattern for decoupling
laymans_terms: Giving a thing its tools instead of letting it find them
analogy: Like hiring a plumber and handing them your wrench set
core_idea: Inversion of control
body: Detailed explanation here
components: Constructor injection, setter injection
why_it_matters: Testability and flexibility
examples: Spring framework, Angular
templates: Use constructor injection by default
tags: design-patterns, oop, testing
---END_ZETTEL---`

    const result = parseZettelSuggestions(input)
    expect(result).toHaveLength(1)
    expect(result[0].title).toBe('Dependency Injection')
    expect(result[0].type).toBe('concept')
    expect(result[0].summary).toBe('A design pattern for decoupling')
    expect(result[0].laymans_terms).toBe('Giving a thing its tools instead of letting it find them')
    expect(result[0].analogy).toBe('Like hiring a plumber and handing them your wrench set')
    expect(result[0].core_idea).toBe('Inversion of control')
    expect(result[0].body).toBe('Detailed explanation here')
    expect(result[0].components).toBe('Constructor injection, setter injection')
    expect(result[0].why_it_matters).toBe('Testability and flexibility')
    expect(result[0].examples).toBe('Spring framework, Angular')
    expect(result[0].templates).toBe('Use constructor injection by default')
    expect(result[0].tags).toEqual(['design-patterns', 'oop', 'testing'])
  })

  it('parses multiple suggestions', () => {
    const input = `---ZETTEL_SUGGESTION---
title: First
type: concept
summary: First summary
---END_ZETTEL---
Some text in between
---ZETTEL_SUGGESTION---
title: Second
type: theory
summary: Second summary
---END_ZETTEL---`

    const result = parseZettelSuggestions(input)
    expect(result).toHaveLength(2)
    expect(result[0].title).toBe('First')
    expect(result[1].title).toBe('Second')
  })

  it('returns empty array for no suggestions', () => {
    expect(parseZettelSuggestions('Just plain text')).toEqual([])
    expect(parseZettelSuggestions('')).toEqual([])
  })

  it('handles missing optional fields with defaults', () => {
    const input = `---ZETTEL_SUGGESTION---
title: Minimal
---END_ZETTEL---`

    const result = parseZettelSuggestions(input)
    expect(result).toHaveLength(1)
    expect(result[0].title).toBe('Minimal')
    expect(result[0].type).toBe('concept')
    expect(result[0].summary).toBe('')
    expect(result[0].body).toBe('')
    expect(result[0].tags).toEqual([])
  })

  it('handles multi-line field content', () => {
    const input = `---ZETTEL_SUGGESTION---
title: Multi
type: concept
body: Line one
  Line two
  Line three
summary: Short
---END_ZETTEL---`

    const result = parseZettelSuggestions(input)
    expect(result).toHaveLength(1)
    expect(result[0].body).toContain('Line one')
    expect(result[0].body).toContain('Line two')
    expect(result[0].body).toContain('Line three')
    expect(result[0].summary).toBe('Short')
  })

  it('parses comma-separated tags', () => {
    const input = `---ZETTEL_SUGGESTION---
title: Tagged
tags: alpha, beta, gamma
---END_ZETTEL---`

    const result = parseZettelSuggestions(input)
    expect(result[0].tags).toEqual(['alpha', 'beta', 'gamma'])
  })
})

describe('parseConnectionSuggestions', () => {
  it('parses connection with all fields', () => {
    const input = `---CONNECTION_SUGGESTION---
source_title: Dependency Injection
target_title: Factory Pattern
label: relates-to
reason: Both deal with object creation
---END_CONNECTION---`

    const result = parseConnectionSuggestions(input)
    expect(result).toHaveLength(1)
    expect(result[0].source_title).toBe('Dependency Injection')
    expect(result[0].target_title).toBe('Factory Pattern')
    expect(result[0].label).toBe('relates-to')
    expect(result[0].reason).toBe('Both deal with object creation')
  })

  it('returns empty array for no connections', () => {
    expect(parseConnectionSuggestions('No connections here')).toEqual([])
    expect(parseConnectionSuggestions('')).toEqual([])
  })

  it('requires source_title and target_title', () => {
    const missingTarget = `---CONNECTION_SUGGESTION---
source_title: Only Source
label: orphan
---END_CONNECTION---`

    const missingSource = `---CONNECTION_SUGGESTION---
target_title: Only Target
label: orphan
---END_CONNECTION---`

    expect(parseConnectionSuggestions(missingTarget)).toEqual([])
    expect(parseConnectionSuggestions(missingSource)).toEqual([])
  })
})
