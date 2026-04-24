# Roadmap

## Done

- [x] `dot init` — create repo directory, git init, .gitignore (`**/.git`, `backups/`)
- [x] `dot add` — move file/directory, create symlink, update index
- [x] `dot restore` — undo tracking, move file back
- [x] `dot install` — recreate symlinks from index on a new machine
- [x] `dot install --overwrite` — backup conflicting files and replace them
- [x] `dot clone <url>` — clone a dotfiles repo and install in one command
- [x] `dot status` — walk the index, report broken symlinks / replaced-by-regular-file
- [x] `dot list` — print all tracked entries from the index
- [x] `dot completion install` — detect shell and install completions automatically
- [x] Portable index with relative paths and file/dir type
- [x] Directory-level symlinks (whole directory as a unit, no per-file default)
- [x] Symlink guards (refuse foreign symlinks, detect already-tracked)
- [x] Conflict backup to `backups/` within the dot repo (gitignored)
- [x] CI pipeline (GitHub Actions: test, build, lint)

## Next

- [ ] Auto-commit on add/restore with `--no-commit` flag to skip
- [ ] `dot packages snapshot` — export `pacman -Qqe` to a tracked file
- [ ] `dot packages install` — install packages from tracked list
- [ ] Goreleaser setup for GitHub Releases
- [ ] Fix `go install github.com/sociale11/dot@latest`

## Later

- [ ] Auto-sync cron setup — `dot cron enable/disable` writes a user crontab entry
- [ ] Auto-push in cron script (with offline tolerance)
- [ ] `dot diff` — show uncommitted changes across all tracked files
- [ ] `dot doctor` — diagnose common issues (broken symlinks, editor safe-write detection, nested repo warnings)
- [ ] Permissions tracking — store file modes in index, restore on install
- [ ] `.dotignore` — exclude patterns within tracked directories
- [ ] Cross-filesystem support — fallback to copy+delete when rename fails
- [ ] Machine tags in index — selective install per hostname

## Maybe

- [ ] Interactive directory picker (fzf-style TUI for selective add)
- [ ] `dot watch` — daemon mode with fsnotify, auto-commit on change
- [ ] Hooks — run scripts before/after add/restore/install
- [ ] Encryption for sensitive configs (SSH, GPG)
