# Contributing

This repository uses different Git author-email rules depending on which remote the resulting commit history is intended for.

## Per-remote author email policy

| Remote | Platform | Author email policy | Why |
|---|---|---|---|
| `origin` | Forgejo | Use default local email | Private Forgejo integration branch keeps workstation-default attribution |
| `gh-fork` | GitHub | Use `-c user.email="n3omod@gmail.com"` on commit-producing commands | GitHub-targeted history should use GitHub attribution email |
| `upstream` | GitHub | Use `-c user.email="n3omod@gmail.com"` on commit-producing commands if direct upstream push is explicitly required | Keeps GitHub-visible attribution consistent |

## Command patterns

### Forgejo `origin`

Use normal Git commands with the default local email:

```bash
git add docs/CONTRIBUTING.md
git commit -m "docs: update contribution guidance"
git push origin main
```

### GitHub `gh-fork`

Use the GitHub email override on commands that create or rewrite commits:

```bash
git add docs/CONTRIBUTING.md
git -c user.email="n3omod@gmail.com" commit -m "docs: update contribution guidance"
git push gh-fork HEAD:my-branch
```

### GitHub history-rewrite commands

If a command will create or rewrite a GitHub-destined commit, keep the same override:

```bash
git -c user.email="n3omod@gmail.com" commit --amend
git -c user.email="n3omod@gmail.com" cherry-pick <commit>
git -c user.email="n3omod@gmail.com" rebase --continue
```

## Important note

`git push` by itself does not change commit author metadata. The email override matters on commands that create or rewrite commits. Apply the correct email policy before the commit exists.
