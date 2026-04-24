# dot

A simple, opinionated dotfiles manager.

## Why?

I want a simple way of managing my dotfiles.

Popular options:

- `stow` simple, but requires a lot of manual work
- `chezmoi` powerful, but requires changing workflow and adding new steps and complexity

**dot** allows you to manage your dotfiles in a central repo, and symlink them back to your original locations.

## How it works

```
dot add ~/.config/nvim    # moves dir to repo, symlinks it back
dot add ~/.zshrc          # same for files
dot restore ~/.zshrc      # undoes tracking, puts the file back
dot install               # on a new machine, recreates all symlinks from the index
```

When you add a file or directory, `dot` moves it into `~/.local/share/dot/`, creates a symlink at the original location, and records the entry in a portable index. The index stores paths relative to `$HOME`, so it works across machines with different usernames.

Directories are symlinked as a unit by default — one symlink for `~/.config/nvim`, not one per file inside it.

Track your dotfiles at `~/.local/share/dot/` with git normally.

## Install

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
dot list

# Check symlink health
dot status

# Push your dotfiles
cd ~/.local/share/dot
git remote add origin git@github.com:<your-username>/dotfiles.git
git push -u origin main

# On a new machine
dot clone git@github.com:<your-username>/dotfiles.git
# or with conflicts
dot clone git@github.com:<your-username>/dotfiles.git --overwrite
```

## Multiple machines
 
The easiest way to manage multiple machines is with git branches. After running `dot clone` on a new machine, branch off and apply any machine-specific changes (e.g. hardware config, monitor layout):
 
```bash
cd ~/.local/share/dot
git checkout -b desktop
# make your changes, commit normally
```
 
If you ever need to reinstall on that machine (e.g. after a disk wipe):
 
```bash
dot clone -b desktop git@github.com:<your-username>/dotfiles.git
```

## Commands

| Command | Description |
|---------|-------------|
| `dot init` | Creates the dot repo at `~/.local/share/dot` and initializes git |
| `dot add <path>...` | Moves files/directories into the repo and symlinks them back |
| `dot restore <path>` | Stops tracking, moves the file back to its original location |
| `dot install` | Reads the index and creates symlinks (for bootstrapping a new machine) |
| `dot install --overwrite` | Backs up conflicting files to `backups/` and replaces them |
| `dot clone <url>` | Clones a dotfiles repo into `~/.local/share/dot` and runs install |
| `dot list` | Prints all tracked entries from the index |
| `dot status` | Shows tracked entries and reports broken or replaced symlinks |
| `dot completion install` | Detects your shell and installs completions |

## Flags

| Flag | Description |
|------|-------------|
| `--root` | Override the root directory (default: `$HOME`) |
| `--dot` | Override the storage directory (default: `~/.local/share/dot`) |

## Design decisions

**Symlinks, not copies.** There's no drift to manage. The file in your repo *is* the file your tools read. No sync step, no apply command.

**Directories as units.** `dot add ~/.config/nvim` creates one symlink for the whole directory. No per-file tracking inside it. If you need to exclude files, use `.gitignore` in the repo.

**Portable index.** Paths are stored relative to `$HOME`. The index doesn't contain `/home/alice/...` — it contains `.config/nvim`. Works on any machine.

**Conflict handling.** `dot install --overwrite` backs up existing files to `backups/` within the dot repo (gitignored) before replacing them. Without `--overwrite`, conflicts are reported and skipped.

## Known tradeoffs

**Editor safe-write.** Some editors (vim by default, vscode sometimes) write to a temp file then rename it over the original. This replaces the symlink with a regular file. Fix for vim: `set backupcopy=yes` in your config. `dot status` will catch broken symlinks.

**Cross-filesystem moves.** `dot add` uses `os.Rename`, which doesn't work across filesystem boundaries. If your home and dot repo are on different mounts, it will fail. This covers the vast majority of setups.

**Nested git repos.** If a tracked directory has its own `.git` (like a nvim config), it becomes a nested repo. The dot repo's `.gitignore` includes `**/.git` to handle this cleanly.

## Project structure

```
dot/
├── cmd/
│   ├── root.go
│   ├── init.go
│   ├── add.go
│   ├── restore.go
│   ├── install.go
│   ├── clone.go
│   ├── status.go
│   ├── list.go
│   ├── completion.go
│   ├── index.go
│   ├── add_test.go
│   ├── index_test.go
│   ├── install_test.go
│   ├── clone_test.go
│   ├── restore_test.go
│   └── status_test.go
├── main.go
├── go.mod
├── go.sum
├── LICENSE
├── README.md
└── ROADMAP.md
```

## Contributing

PRs welcome. Run tests before submitting:

```bash
go test ./...
```

## License

MIT

