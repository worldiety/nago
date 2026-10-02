---
title: Configuration and Deployment
linkTitle: Deployment
weight: 9
---

## Build

A Nago application is a regular Go program. `go build` produces a single executable which contains the server,
the embedded frontend and all assets. Nago uses no cgo, so you can cross-compile, e.g. for a Linux server:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp .
```

At runtime the binary needs nothing but a writable [data directory](../persistence/#the-data-directory).

## Environment variables

Nago reads the following variables at startup. `application.Configure` also loads a `.env` file from the working
directory, if present; variables which are already set in the environment take precedence.

| Variable                 | Default                    | Purpose                                                                                     |
|--------------------------|----------------------------|---------------------------------------------------------------------------------------------|
| `HOST`                   | `localhost`                | address the server binds to; use `0.0.0.0` in containers or behind a reverse proxy on another host |
| `PORT`                   | `3000`                     | port of the HTTP server                                                                     |
| `HOSTNAME`               | taken from the first request | public DNS name; Nago uses `https://<HOSTNAME>` to generate absolute links, e.g. in mails  |
| `STATE_DIRECTORY`        | `~/.nago/<application ID>` | data directory, set by systemd for services with `StateDirectory=`                          |
| `NAGO_MASTER_KEY`        | generated file             | hex-encoded 32 byte key to encrypt sensitive data, see [Master key](#master-key)            |
| `NO_SSL`                 | depends on the OS, see [Debug mode](#debug-mode) | `true` marks the session cookie as not secure, `false` as secure                            |

`NAGO_HOST` and `NAGO_PORT` are deprecated aliases of `HOST` and `PORT`.

The variables are applied before your configuration function runs, so calls like `cfg.SetHost` or
`cfg.SetDataDir` in your code override them.

{{< callout type="warning" >}}
Some container runtimes, e.g. Docker, set `HOSTNAME` to the container ID. Set it explicitly to your public DNS
name, otherwise generated links point to the wrong host.
{{< /callout >}}

At startup Nago logs all environment variables. Values of variables whose names hint at secrets, like `KEY`,
`TOKEN` or `PASS`, and passwords within URLs are masked.

## Debug mode

The debug mode is on by default on macOS and Windows and off on all other systems; change it with
`cfg.Debug(bool)`. In debug mode Nago

- logs human-readable text instead of JSON,
- allows cross-origin requests, which the frontend development server needs,

The session cookie is decided separately, before your configuration runs: without `NO_SSL`, it is non-secure on
macOS and Windows, so that login works over plain `http://localhost`, and secure on all other systems. A later
`cfg.Debug` call does not change it.

In production, i.e. on Linux, cookies are secure and therefore require HTTPS. Terminate TLS in a reverse proxy in
front of the application. The proxy must forward websocket connections.

## Master key

Secrets, sessions and other sensitive data are encrypted with a 32 byte master key. If `NAGO_MASTER_KEY` is not
set, Nago generates a random key on the first start and stores it in `.masterkey` in the data directory. Anyone
who can read the data directory can therefore decrypt it; for production, provide the key from your secret
management:

```bash
export NAGO_MASTER_KEY=$(openssl rand -hex 32)
```

Keep the key safe and separate from your backups. Without it, the encrypted parts of a backup cannot be restored.
The [backup system](/docs/systems/backup_management/) can export and replace the key.

## Running as a systemd service

```ini
[Service]
ExecStart=/opt/myapp/myapp
DynamicUser=yes
StateDirectory=myapp
Environment=HOST=127.0.0.1 PORT=3000 HOSTNAME=myapp.example.com
EnvironmentFile=/etc/myapp/secrets.env
Restart=on-failure
```

systemd creates `/var/lib/myapp` and passes it as `STATE_DIRECTORY`, which Nago uses as data directory as is,
without appending the application ID. On `systemctl stop`, Nago receives `SIGTERM` and shuts down gracefully.

## Reset the admin password

If nobody can log in anymore, enable the bootstrap admin `admin@localhost` with the `nago-adm` tool. Stop the
application first, because the database is locked while it runs:

```bash
go run go.wdy.de/nago/cmd/nago-adm@latest -app=com.example.myapp -cmd=admin-reset -pwd='<new password>' -lifetime=1h
```

Pass `-data-dir=<dir>` if the data directory is not `~/.nago/<application ID>`, e.g. the directory of
`STATE_DIRECTORY`. After the next start, log in as `admin@localhost` and fix the accounts. The account expires
after the given lifetime, counted from the moment you run the tool; `-lifetime=0` keeps it enabled
indefinitely.

## Operating checklist

- A written license agreement with worldiety, see [Installation](../../getting-started/installation/#license).
- `HOST`, `PORT` and `HOSTNAME` set, TLS terminated in a reverse proxy with websocket support.
- A persistent data directory and a backup of it.
- `NAGO_MASTER_KEY` set and stored separately from the backups.
- No hard-coded bootstrap admin passwords in the code.
