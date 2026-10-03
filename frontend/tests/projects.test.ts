import { describe, expect, it } from 'vitest'
import { allProjects, featuredProjects, getProject, projects } from '~/content/projects'
import { isPending } from '~/utils/content'

describe('project content integrity', () => {
  it('has a unique, url-safe slug for every project', () => {
    const slugs = projects.map((p) => p.slug)
    expect(new Set(slugs).size).toBe(slugs.length)
    for (const slug of slugs) expect(slug).toMatch(/^[a-z0-9]+(-[a-z0-9]+)*$/)
  })

  it('never publishes a placeholder as a link (§27)', () => {
    for (const project of projects) {
      for (const url of [project.liveUrl, project.githubUrl]) {
        if (url === undefined) continue
        expect(isPending(url), `${project.slug} has a placeholder URL`).toBe(false)
        expect(url).toMatch(/^https:\/\//)
      }
    }
  })

  it('gives every project a title and a category', () => {
    for (const project of projects) {
      expect(project.title.trim()).not.toBe('')
      expect(project.category.trim()).not.toBe('')
    }
  })

  it('never lists a placeholder as a technology', () => {
    // An empty list is fine — it means the stack is not known yet, and the tag
    // row simply does not render. A `TODO:` entry masquerading as a technology
    // would not be fine.
    for (const project of projects) {
      for (const tech of project.technologies) {
        expect(isPending(tech), `${project.slug} has a placeholder technology`).toBe(false)
        expect(tech.trim()).not.toBe('')
      }
    }
  })

  it('orders featured projects by their explicit order', () => {
    const order = featuredProjects.map((p) => p.order ?? 99)
    expect(order).toEqual([...order].sort((a, b) => a - b))
  })

  it('includes every featured project in the full list', () => {
    for (const project of featuredProjects) {
      expect(allProjects.map((p) => p.slug)).toContain(project.slug)
    }
  })

  it('looks a project up by slug and returns undefined for an unknown one', () => {
    expect(getProject('portfolio')?.title).toBe('Portfolio')
    expect(getProject('does-not-exist')).toBeUndefined()
  })
})
