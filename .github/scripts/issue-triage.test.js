const { describe, it } = require('node:test');
const assert = require('node:assert/strict');
const { parseComment } = require('./issue-triage.js');

const known = [
  'status:available',
  'status:claimed',
  'status:in-progress',
  'status:review',
  'status:completed',
  'module:backend',
];

function parse(body, extra = {}) {
  return parseComment(body, { login: 'alice', known, isPR: false, isOpen: true, assignees: [], ...extra });
}

describe('parseComment phrases', () => {
  it('认领未分配的 Issue', () => {
    const r = parse('我来认领');
    assert.deepEqual([...r.toAssign], ['alice']);
    assert.ok(r.toAdd.has('status:claimed'));
    assert.ok(r.toRemove.has('status:available'));
    assert.equal(r.notice, '');
  });

  it('已有经办人时拒绝认领', () => {
    const r = parse('我来认领', { assignees: ['bob'] });
    assert.deepEqual([...r.toAssign], []);
    assert.equal(r.toAdd.size, 0);
    assert.match(r.notice, /已分配给 @bob/);
  });

  it('评论里提到认领但不独立成行则忽略', () => {
    const r = parse('我想说我来认领这件事，但还没决定');
    assert.equal(r.toAssign.size, 0);
    assert.equal(r.toAdd.size, 0);
  });

  it('只有经办人能放弃认领 / 开始开发 / 提交审查', () => {
    const drop = parse('放弃认领', { assignees: ['bob'] });
    assert.equal(drop.toUnassign.size, 0);
    assert.equal(drop.toAdd.size, 0);

    const start = parse('开始开发', { assignees: ['alice'] });
    assert.ok(start.toAdd.has('status:in-progress'));

    const review = parse('提交审查', { assignees: ['alice'] });
    assert.ok(review.toAdd.has('status:review'));
  });

  it('关闭的 Issue 不响应口头命令', () => {
    const r = parse('我来认领\n开始开发\n提交审查', { isOpen: false });
    assert.equal(r.toAssign.size, 0);
    assert.equal(r.toAdd.size, 0);
  });

  it('PR 评论不走认领短语', () => {
    const r = parse('我来认领', { isPR: true });
    assert.equal(r.toAssign.size, 0);
  });

  it('/label 仍可改已有标签', () => {
    const r = parse('/label module:backend');
    assert.ok(r.toAdd.has('module:backend'));
  });
});
