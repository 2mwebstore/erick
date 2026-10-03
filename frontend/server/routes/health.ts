/**
 * Frontend liveness endpoint (§33).
 *
 * Reports only on this process. It deliberately does not proxy to the Go API's
 * /health: an uptime monitor should be able to tell "the site is down" apart
 * from "the contact form is down", and collapsing them into one signal loses
 * that distinction.
 */
export default defineEventHandler((event) => {
  setHeader(event, 'cache-control', 'no-store')

  return {
    status: 'ok',
    service: 'portfolio-web',
    time: new Date().toISOString(),
  }
})
