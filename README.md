# dot

A minimal dotfile tracker. No copying, no templating, no sync commands — just symlinks and git.

You edit your config files normally. `dot` moves them into a central repo and symlinks them back. Changes flow through the symlinks transparently. Your repo stays in sync because it *is* the source of truth.

## How it works

```
dot add ~/.config/nvim    # moves dir to repo, symlinks it back
dot add ~/.zshrc          # same for files
dot restore ~/.zshrc      # undoes tracking, puts the file back
dot install               # on a new machine, recreates all symlinks from the index
```

When you add a file or directory, `dot` moves it into `~/.local/share/dot/`, creates a symlink at the original location, and records the entry in a portable index. The index stores paths relative to `$HOME`, so it works across machines with different usernames.

Directories are symlinked as a unit by default — one symlink for `~/.config/nvim`, not one per file inside it.

## Install

```bash
go install github.com/sociale11/dot@latest
```

Or build from source:

```bash
git clone https://github.com/sociale11/dot.git
cd dot
go build -o dot .
```

## Quick start

```bash
# Initialize the repo
dot init

# Start tracking
dot add ~/.zshrc
dot add ~/.config/nvim
dot add ~/.config/git/config

# Check what's tracked
dot status

# Push your dotfiles
cd ~/.local/share/dot
git remote add origin git@github.com:sociale11/dotfiles.git
git push -u origin main

# On a new machine
git clone git@github.com:sociale11/dotfiles.git ~/.local/share/dot
dot install
```

## Commands

| Command | Description |
|---------|-------------|
| `dot init` | Creates the dot repo at `~/.local/share/dot` and initializes git |
| `dot add <path>...` | Moves files/directories into the repo and symlinks them back |
| `dot restore <path>` | Stops tracking, moves the file back to its original location |
| `dot install` | Reads the index and creates symlinks (for bootstrapping a new machine) |
| `dot status` | Shows tracked entries and their health |

## Flags

| Flag | Description |
|------|-------------|
| `--root` | Override the root directory (default: `$HOME`) |
| `--dot` | Override the storage directory (default: `~/.local/share/dot`) |
| `--no-commit` | Skip the auto git commit after add/restore |

## Design decisions

**Symlinks, not copies.** There's no drift to manage. The file in your repo *is* the file your tools read. No sync step, no apply command.

**Directories as units.** `dot add ~/.config/nvim` creates one symlink for the whole directory. No per-file tracking inside it. If you need to exclude files, use `.gitignore` in the repo.

**Portable index.** Paths are stored relative to `$HOME`. The index doesn't contain `/home/alice/...` — it contains `.config/nvim`. Works on any machine.

**Auto-commit.** Each `add` and `restore` creates a git commit. Skip with `--no-commit` if you want to batch changes.

## Auto-sync edits

When you edit a tracked file, the change lands in the dot repo through the symlink — but nothing commits it. A cron job handles this:

```bash
# Add to your crontab: crontab -e
*/5 * * * * cd ~/.local/share/dot && git add -A && git diff --cached --quiet || git commit -m "auto: $(date +\%Y-\%m-\%d\ \%H:\%M)"
```

## Known tradeoffs

**Editor safe-write.** Some editors (vim by default, vscode sometimes) write to a temp file then rename it over the original. This replaces the symlink with a regular file. Fix for vim: `set backupcopy=yes` in your config. `dot status` will catch broken symlinks.

**Cross-filesystem moves.** `dot add` uses `os.Rename`, which doesn't work across filesystem boundaries. If your home and dot repo are on different mounts, it will fail. This covers the vast majority of setups.

**Nested git repos.** If a tracked directory has its own `.git` (like a nvim config), it becomes a nested repo. The dot repo's `.gitignore` includes `**/.git` to handle this cleanly.

## Project structure

```
dot/
├── cmd/
│   ├── root.go          # cobra root command, global flags
│   ├── init.go          # dot init
│   ├── add.go           # dot add
│   ├── restore.go       # dot restore
│   ├── install.go       # dot install
│   ├── index.go         # index read/write logic
│   ├── add_test.go
│   └── index_test.go
├── main.go
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Contributing

PRs welcome. Run tests before submitting:

```bash
go test ./...
```

## License

MIT
