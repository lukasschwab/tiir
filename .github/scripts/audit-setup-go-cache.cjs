const assert = require('node:assert/strict');
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');

const repo = process.env.GITHUB_REPOSITORY;
const run = process.env.GITHUB_RUN_ID;
const attempt = process.env.GITHUB_RUN_ATTEMPT;
const prefix = `modes-${run}-${attempt}-`;
const api = endpoint => execFileSync('gh', ['api', endpoint], { encoding: 'utf8' });
const pages = endpoint => JSON.parse(execFileSync('gh', [
  'api', '--paginate', '--slurp', endpoint,
], { encoding: 'utf8' }));

const caches = pages(`repos/${repo}/actions/caches?per_page=100`)
  .flatMap(page => page.actions_caches)
  .filter(cache => cache.key.includes(`-${prefix}`));
for (const mode of ['default', 'true', 'false', 'failure']) {
  const count = caches.filter(cache => cache.key.includes(`-${prefix}${mode}-`)).length;
  assert.equal(count, ['default', 'true'].includes(mode) ? 2 : 0,
    `Unexpected number of uploaded caches for ${mode}`);
}

const jobs = pages(`repos/${repo}/actions/runs/${run}/attempts/${attempt}/jobs?per_page=100`)
  .flatMap(page => page.jobs)
  .filter(job => job.name.startsWith('seed-') || job.name.startsWith('restore-'));
assert.equal(jobs.length, 12);
const timings = [];
for (const job of jobs) {
  const intentionalFailure = job.name === 'seed-failure';
  assert.equal(job.conclusion, intentionalFailure ? 'failure' : 'success', job.name);
  const log = api(`repos/${repo}/actions/jobs/${job.id}/logs`);
  const shouldSave = ['seed-default', 'seed-true',
    'restore-default-prefix-save', 'restore-true-prefix-save'].includes(job.name);
  const shouldRegister = shouldSave || intentionalFailure;
  const registration = /##\[group\]Run actions\/cache@/.test(log);
  assert.equal(registration, shouldRegister, `${job.name}: save registration`);
  assert.equal(log.includes('Cache saved with key:'), shouldSave, `${job.name}: upload`);
  if (shouldRegister) {
    assert.ok(log.includes('lookup-only: true'), `${job.name}: metadata-only lookup`);
  }
  if (shouldSave) {
    const trimmed = log.indexOf('GOCACHE trim:');
    const saved = log.indexOf('Cache saved with key:');
    assert.ok(trimmed >= 0 && trimmed < saved, `${job.name}: trim before upload`);
  }
  const lines = log.split('\n');
  const start = lines.find(line => line.includes('##[start-action display=register cache save;'));
  if (start) {
    const id = start.match(/;id=([^\]]+)/)[1];
    const end = lines.find(line => line.includes(`##[end-action id=${id};`));
    assert.ok(end, `${job.name}: registration completed`);
    const duration = Number(end.match(/duration_ms=(\d+)/)[1]);
    timings.push([job.name, duration]);
  }
  console.log(`PASS ${job.name}: register=${registration}, upload=${shouldSave}`);
}

const summary = [
  '## Setup Go cache mode validation',
  '',
  'All cold-miss, default/explicit-save, restore-only, prefix-refresh, exact-hit,',
  'failed-job, trim-before-upload, and restored-fixture checks passed.',
  '',
  '| Save-registration step | Duration (ms) |',
  '| --- | ---: |',
  ...timings.map(([name, ms]) => `| ${name} | ${ms} |`),
  '',
  `Verified ${caches.length} uploaded entries; restore-only and failed jobs uploaded none.`,
  '',
].join('\n');
fs.appendFileSync(process.env.GITHUB_STEP_SUMMARY, summary);
console.log(summary);
