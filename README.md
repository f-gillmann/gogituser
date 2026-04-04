# gogituser

Switch git identities per repo. Profiles are stored in `~/.gituser.toml` and applied via `git config` using your configured commit name, email, GPG signing key, and SSH key.

## Install

```sh
just install
```

## Usage

```sh
gogituser add                           # create a profile
gogituser use <profile>                 # apply to current repo
gogituser list                          # list all profiles
gogituser update <profile>              # edit a profile
gogituser delete <profile>              # remove a profile
gogituser clone <profile> <repo>        # clone using profile's SSH key
gogituser remote set-url <profile>      # rewrite existing remote URL to use SSH host alias
gogituser remote add <profile> <url>    # add a new remote with SSH host alias applied
```

`use`, `remote set-url`, and `remote add` must be run inside a git repository.

## Config

Profiles are stored in `~/.gituser.toml`:

```toml
[[profile]]
name        = "personal"
git_name    = "Your Name"
email       = "you@example.com"
signing_key = "ABCD1234"
ssh_key     = "~/.ssh/id_personal"
ssh_host    = "github-personal"     # optional, see below
```

## Multiple accounts on the same host

To use two accounts on the same host (e.g. two GitHub accounts), set up SSH host aliases in `~/.ssh/config`:

```
Host github-personal
    HostName        github.com
    User            git
    IdentityFile    ~/.ssh/id_personal
    AddKeysToAgent  yes

Host github-work
    HostName        github.com
    User            git
    IdentityFile    ~/.ssh/id_work
    AddKeysToAgent  yes
```

Then set `ssh_host` on the profile to the alias name and use `remote set-url` in existing repos or `remote add` for new ones:

```sh
gogituser remote set-url personal           # rewrites origin: git@github.com:... -> git@github-personal:...
gogituser remote set-url personal upstream  # target a different remote
gogituser remote add personal <url>         # add a new remote to your local repo
```

`remote set-url` and `remote add` warn if the alias is missing from `~/.ssh/config` or if the file doesn't exist yet.
