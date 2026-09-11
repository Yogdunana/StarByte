function parseComment(body, { login, known, isPR, assignees = [], isOpen = true }) {
  const knownSet = known instanceof Set ? known : new Set(known);
  const toAdd = new Set();
  const toRemove = new Set();
  const toAssign = new Set();
  const toUnassign = new Set();
  const unknown = [];

  function matchKnown(text) {
    const trimmed = text.trim();
    if (!trimmed) return null;
    const sorted = [...knownSet].sort((a, b) => b.length - a.length);
    return sorted.find((name) => trimmed === name) || null;
  }

  function takeLabels(text) {
    let remaining = text.trim();
    const names = [];
    while (remaining) {
      const sorted = [...knownSet].sort((a, b) => b.length - a.length);
      const hit = sorted.find((name) => remaining === name || remaining.startsWith(`${name} `));
      if (hit) {
        names.push(hit);
        remaining = remaining.slice(hit.length).trim();
        continue;
      }
      const tok = remaining.split(/\s+/)[0];
      unknown.push(tok);
      remaining = remaining.slice(tok.length).trim();
    }
    return names;
  }

  function mentions(text) {
    const users = [];
    const re = /@([A-Za-z0-9-]+)/g;
    let m;
    while ((m = re.exec(text))) {
      const name = m[1];
      if (/^me$/i.test(name)) users.push(login);
      else users.push(name);
    }
    if (/(^|\s)me(\s|$)/i.test(text.replace(/@me\b/gi, ' '))) users.push(login);
    return [...new Set(users)];
  }

  function addStatus(name) {
    for (const label of knownSet) {
      if (label.startsWith('status:') && label !== name) toRemove.add(label);
    }
    toAdd.add(name);
    toRemove.delete(name);
  }

  const commandRe =
    /^\s*\/(label|unlabel|assign|unassign|加标签|去掉标签|取消标签|移除标签|指派|取消指派)\s+(.+)\s*$/gim;
  let cmd;
  while ((cmd = commandRe.exec(body))) {
    const kind = cmd[1].toLowerCase();
    const rest = cmd[2];
    if (kind === 'label' || kind === '加标签') {
      for (const name of takeLabels(rest)) {
        if (name.startsWith('status:')) addStatus(name);
        else toAdd.add(name);
      }
    } else if (kind === 'unlabel' || kind === '去掉标签' || kind === '取消标签' || kind === '移除标签') {
      for (const name of takeLabels(rest)) toRemove.add(name);
    } else if (kind === 'assign' || kind === '指派') {
      const users = mentions(rest);
      if (!users.length) toAssign.add(login);
      else users.forEach((u) => toAssign.add(u));
    } else if (kind === 'unassign' || kind === '取消指派') {
      const users = mentions(rest);
      if (!users.length) toUnassign.add(login);
      else users.forEach((u) => toUnassign.add(u));
    }
  }

  for (const line of body.split(/\r?\n/)) {
    const plus = line.match(/^\s*\+\s*(.+?)\s*$/);
    const minus = line.match(/^\s*-\s*(.+?)\s*$/);
    if (plus) {
      const name = matchKnown(plus[1]);
      if (!name) continue;
      if (name.startsWith('status:')) addStatus(name);
      else toAdd.add(name);
    } else if (minus) {
      const name = matchKnown(minus[1]);
      if (name) toRemove.add(name);
    }
  }

  const isAssignee = assignees.includes(login);
  const isOpenIssue = !isPR && isOpen;
  let notice = '';

  function hasPhrase(re) {
    return body.split(/\r?\n/).some((line) => re.test(line));
  }

  if (isOpenIssue && hasPhrase(/^\s*我来认领/)) {
    if (assignees.length && !isAssignee) {
      notice = `@${login} 该 Issue 已分配给 ${assignees.map((a) => '@' + a).join('、')}，请先沟通再认领。`;
    } else {
      toAssign.add(login);
      addStatus('status:claimed');
    }
  }
  if (isOpenIssue && isAssignee && hasPhrase(/^\s*放弃认领/)) {
    toUnassign.add(login);
    addStatus('status:available');
  }
  if (isOpenIssue && isAssignee && hasPhrase(/^\s*开始开发/)) addStatus('status:in-progress');
  if (isOpenIssue && isAssignee && hasPhrase(/^\s*(?:提交审查|提交评审)/)) addStatus('status:review');

  for (const name of toRemove) toAdd.delete(name);

  return { toAdd, toRemove, toAssign, toUnassign, unknown, notice };
}

async function issueTriage({ github, context }) {
  const body = context.payload.comment.body || '';
  const login = context.payload.comment.user.login;
  const issue = context.payload.issue;
  const owner = context.repo.owner;
  const repo = context.repo.repo;
  const issue_number = issue.number;
  const isPR = Boolean(issue.pull_request);

  const known = new Set(
    (await github.paginate(github.rest.issues.listLabelsForRepo, { owner, repo, per_page: 100 })).map(
      (label) => label.name,
    ),
  );

  const currentLabels = (issue.labels || []).map((l) => l.name);
  const assignees = (issue.assignees || []).map((a) => a.login);
  const isOpen = issue.state === 'open';
  const { toAdd, toRemove, toAssign, toUnassign, unknown, notice } = parseComment(body, {
    login,
    known,
    isPR,
    assignees,
    isOpen,
  });
  const errors = [];

  if (notice) {
    await github.rest.issues.createComment({ owner, repo, issue_number, body: notice });
  }

  if (!toAdd.size && !toRemove.size && !toAssign.size && !toUnassign.size) {
    if (unknown.length) {
      await github.rest.issues.createComment({
        owner,
        repo,
        issue_number,
        body: `@${login} 这些标签不存在：${unknown.map((n) => `\`${n}\``).join('、')}。请用仓库里已有的 label。`,
      });
    }
    return;
  }

  for (const name of toAdd) {
    if (!currentLabels.includes(name)) {
      try {
        await github.rest.issues.addLabels({ owner, repo, issue_number, labels: [name] });
      } catch (err) {
        errors.push(`加标签 \`${name}\` 失败：${err.message}`);
      }
    }
  }

  for (const name of toRemove) {
    if (!currentLabels.includes(name) && !toAdd.has(name)) continue;
    try {
      await github.rest.issues.removeLabel({ owner, repo, issue_number, name });
    } catch (err) {
      if (err.status !== 404) errors.push(`去掉标签 \`${name}\` 失败：${err.message}`);
    }
  }

  if (toAssign.size) {
    try {
      await github.rest.issues.addAssignees({
        owner,
        repo,
        issue_number,
        assignees: [...toAssign],
      });
    } catch (err) {
      errors.push(`指派 ${[...toAssign].map((u) => '@' + u).join(' ')} 失败：${err.message}`);
    }
  }

  if (toUnassign.size) {
    try {
      await github.rest.issues.removeAssignees({
        owner,
        repo,
        issue_number,
        assignees: [...toUnassign],
      });
    } catch (err) {
      errors.push(`取消指派 ${[...toUnassign].map((u) => '@' + u).join(' ')} 失败：${err.message}`);
    }
  }

  const notes = [];
  if (unknown.length) notes.push(`这些标签不存在：${unknown.map((n) => `\`${n}\``).join('、')}`);
  notes.push(...errors);
  if (notes.length) {
    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number,
      body: `@${login} ${notes.join('\n')}`,
    });
  }
}

module.exports = issueTriage;
module.exports.parseComment = parseComment;
