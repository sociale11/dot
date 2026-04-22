# Roadmap

## Done

- [x] `dot init` — create repo directory, git init, .gitignore
- [x] `dot add` — move file/directory, create symlink, update index
- [x] `dot restore` — undo tracking, move file back
- [x] `dot install` — recreate symlinks from index on a new machine
- [x] Portable index with relative paths
- [x] Directory-level symlinks (no per-file recursion by default)
- [x] Symlink guards (refuse foreign symlinks, detect already-tracked)

## Next

- [x] `dot status` — walk the index, report broken symlinks / replaced-by-regular-file / untracked changes
- [ ] Auto-commit on add/restore with `--no-commit` flag
- [x] Git init + .gitignore (`**/.git`) during `dot init`
- [ ] Conflict handling on `dot install` — show diff when target exists, offer skip/overwrite/backup
- [x] `dot list` — print all tracked entries from the index

## Later

- [ ] Auto-sync cron setup — `dot cron enable/disable` writes a user crontab entry
- [ ] Auto-push in cron script (with offline tolerance)
- [ ] `dot diff` — show uncommitted changes across all tracked files
- [ ] Permissions tracking — store file modes in index, restore on install
- [ ] `.dotignore` — exclude patterns within tracked directories
- [ ] `dot doctor` — diagnose common issues (broken symlinks, editor safe-write detection, nested repo warnings)
- [ ] Cross-filesystem support — fallback to copy+delete when rename fails
- [ ] Shell completions (bash, zsh, fish)

## Maybe

- [ ] Interactive directory picker (fzf-style TUI for selective add)
- [ ] `dot watch` — daemon mode with fsnotify, auto-commit on change
- [ ] Hooks — run scripts before/after add/restore/install
- [ ] Per-machine branches or tags
- [ ] Encryption for sensitive configs (SSH, GPG)
