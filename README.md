# gogituser

`ggu` switches git identities per repo. A profile holds your commit name, email, and optionally a GPG signing key and an SSH key. `ggu use` writes them into the repo's git config, so plain `git` uses the right identity from then on.

## Build

```sh
just build          # binary in bin/ggu
just completions    # bash, zsh and fish completions in bin/completions/
```

Run `just` to see all recipes.

## Usage

```sh
ggu add [profile]             # create a profile
ggu use [profile]             # apply a profile to the current repo, opens a picker without a name
ggu list                      # list all profiles, marks the one active in this repo
ggu update <profile>          # change a profile
ggu remove <profile>          # remove a profile
ggu git <profile> <args...>   # run any git command as a profile
```

`add` and `update` open a form, and `use` without a name opens a picker.

## Without a terminal

Every command also works in scripts, CI or dotfile setups. Pass the values as flags and no form opens:

```sh
ggu add work --git-name "Your Name" --git-email you@work.com \
             --signing-key 1989EF9FE4BBD03D --ssh-key ~/.ssh/id_work
ggu update work --git-email new@work.com     # changes only this field
ggu update work --signing-key ""             # clears an optional field
ggu use work
ggu remove work --yes
```

Without a terminal, ggu never opens a form. It tells you which flags are missing instead.

## What `use` sets

| Profile field | git config |
|---|---|
| Git Name | `user.name` |
| Git Email | `user.email` |
| Signing Key | `user.signingkey`, plus `commit.gpgsign` and `tag.gpgsign` |
| SSH Key | `core.sshCommand = ssh -i <key> -o IdentitiesOnly=yes` |

Empty fields are removed from the repo config, so nothing from a previous profile stays active. ggu also writes `ggu.profile` so `list` knows which profile is active.

## Running git as a profile

`ggu git <profile>` runs git with the profile's identity. Anything after the profile name goes straight to git:

```sh
ggu git personal clone git@github.com:me/repo.git
ggu git personal push
ggu git work fetch upstream
```

After `clone`, the profile is applied to the new repo, so you can keep using plain `git` there.

## Config

Profiles are stored in `~/.config/ggu/config.toml` (or `$XDG_CONFIG_HOME/ggu/config.toml`):

```toml
[[profile]]
name        = "personal"
git_name    = "Your Name"
git_email   = "you@example.com"
signing_key = "1989EF9FE4BBD03D"       # optional, turns on commit and tag signing
ssh_key     = "~/.ssh/id_personal"     # optional, without it ssh and your agent pick the key
```

SSH key paths must be full paths (`/...` or `~/...`). Paths in your home folder are saved with `~/`, so the config works on other machines too.

## Signing and GitHub

GitHub only marks a signed commit as verified when the commit email is also on the GPG key. `ggu list` warns when they differ, and the form marks keys with another email.

## Multiple accounts on the same host

Give each account its own SSH key and set `ssh_key` on its profile. ggu then tells ssh to use only that key, so GitHub knows which account you are. Remote URLs stay as they are, like `git@github.com:me/repo.git`.

```sh
ggu git work clone git@github.com:company/repo.git   # clones with the work key
ggu use personal                                      # this repo now pushes with the personal key
```
