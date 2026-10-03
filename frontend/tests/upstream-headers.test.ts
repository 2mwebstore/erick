// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { pickUpstreamHeaders } from '~/server/utils/upstream-headers'

/**
 * What a sign-in request looks like by the time it reaches the Nuxt server on
 * Railway: the browser's own headers plus everything Cloudflare and Railway's
 * edge added on the way in.
 */
const incoming = {
  'accept': 'application/json',
  'content-type': 'application/json',
  'cookie': 'portfolio_session=abc; portfolio_csrf=def',
  'user-agent': 'Mozilla/5.0',
  'x-csrf-token': 'def',
  'host': 'kongchansila.com',
  'cf-connecting-ip': '203.0.113.7',
  'cf-ray': '8c0e1a2b3c4d5e6f-HKG',
  'cf-visitor': '{"scheme":"https"}',
  'cf-ipcountry': 'KH',
  'cdn-loop': 'cloudflare',
  'x-forwarded-for': '203.0.113.7, 172.70.1.1',
  'x-forwarded-proto': 'https',
  'x-real-ip': '172.70.1.1',
  'x-railway-edge': 'railway/asia-southeast1',
}

describe('pickUpstreamHeaders', () => {
  it('forwards what the API reads: session, CSRF, body type and user agent', () => {
    const headers = pickUpstreamHeaders(incoming, '203.0.113.7')
    expect(headers).toEqual({
      'accept': 'application/json',
      'content-type': 'application/json',
      'cookie': 'portfolio_session=abc; portfolio_csrf=def',
      'user-agent': 'Mozilla/5.0',
      'x-csrf-token': 'def',
      'x-real-ip': '203.0.113.7',
    })
  })

  // Forwarding these back into Cloudflare is what made it answer 403
  // "DNS points to prohibited IP" to every sign-in.
  it('drops every header Cloudflare or the edge added', () => {
    const forwarded = Object.keys(pickUpstreamHeaders(incoming, '203.0.113.7'))
    for (const name of ['host', 'cf-connecting-ip', 'cf-ray', 'cf-visitor', 'cf-ipcountry', 'cdn-loop', 'x-forwarded-for', 'x-forwarded-proto', 'x-railway-edge']) {
      expect(forwarded).not.toContain(name)
    }
  })

  it('sets x-real-ip from the visitor, never from what the client sent', () => {
    expect(pickUpstreamHeaders({ 'x-real-ip': '10.0.0.1' }, '203.0.113.7')['x-real-ip']).toBe('203.0.113.7')
    expect(pickUpstreamHeaders({ 'x-real-ip': '10.0.0.1' }, '')).not.toHaveProperty('x-real-ip')
  })

  it('matches header names in any case', () => {
    expect(pickUpstreamHeaders({ 'Cookie': 'a=1', 'X-CSRF-Token': 't' }, '')).toEqual({
      'cookie': 'a=1',
      'x-csrf-token': 't',
    })
  })
})
